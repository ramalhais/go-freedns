package freedns

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type ConfigUrls struct {
	Base         string `envconfig:"URLS_BASE"`
	Login        string `envconfig:"URLS_LOGIN"`
	GetDomains   string `envconfig:"URLS_GET_DOMAINS"`
	CreateDomain string `envconfig:"URLS_CREATE_DOMAIN"`
	DeleteDomain string `envconfig:"URLS_DELETE_DOMAIN"`
	GetRecords   string `envconfig:"URLS_GET_RECORDS"`
	UpdateRecord string `envconfig:"URLS_UPDATE_RECORD"`
	DeleteRecord string `envconfig:"URLS_DELETE_RECORD"`
}

type ConfigAuth struct {
	Login       string `envconfig:"AUTH_LOGIN"`
	Password    string `envconfig:"AUTH_PASSWORD"`
	CookieName  string `envconfig:"AUTH_COOKIE_NAME"`
	CookieValue string `envconfig:"AUTH_COOKIE_VALUE"`
}

type FreeDNS struct {
	Urls   ConfigUrls
	Auth   ConfigAuth
	Client *http.Client
}

func (ctx *FreeDNS) ConfigFile() error {
	f, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (ctx *FreeDNS) ConfigEnv() error {
	err := envconfig.Process("", ctx)
	if err != nil {
		return err
	}

	return nil
}

func (ctx *FreeDNS) Authenticate() (string, error) {
	if ctx.Auth.Login == "" || ctx.Auth.Password == "" {
		return "", errors.New("Auth not found in configuration")
	}
	formData := url.Values{
		"username": []string{ctx.Auth.Login},
		"password": []string{ctx.Auth.Password},
		"action":   []string{"auth"},
	}
	resp, err := ctx.Client.PostForm(ctx.Urls.Base+ctx.Urls.Login, formData)
	if err != nil {
		return "", err
	}
	cookie, err := resp.Request.Cookie(ctx.Auth.CookieName)
	ctx.Auth.CookieValue = cookie.Value

	return ctx.Auth.CookieValue, err
}

func NewFreeDNS() (*FreeDNS, error) {
	ctx := &FreeDNS{
		Urls: ConfigUrls{
			Base:         "https://freedns.afraid.org",
			Login:        "/zc.php?step=2",
			GetDomains:   "/domain/",
			CreateDomain: "/domain/domaincheck.php?domain={DOMAIN}",
			DeleteDomain: "/domain/delete.php?domain_id={DOMAIN_ID}",
			GetRecords:   "/subdomain/?limit={DOMAIN_ID}",
			UpdateRecord: "/subdomain/save.php?step=2",
			DeleteRecord: "/subdomain/delete2.php?data_id%5B%5D={RECORD_ID}&submit=delete+selected",
		},
		Auth: ConfigAuth{
			CookieName: "dns_cookie",
		},
	}

	var err error
	err = ctx.ConfigFile()
	if err != nil {
		fmt.Printf("ConfigFile: %s\n", err)
	}
	err = ctx.ConfigEnv()
	if err != nil {
		fmt.Printf("Error: ConfigEnv: %s", err)
	}

	jar, _ := cookiejar.New(nil)
	ctx.Client = &http.Client{
		Jar: jar,
	}

	_, err = ctx.Authenticate()

	return ctx, err
}

func (ctx *FreeDNS) GetDomains() (map[string]string, error) {
	resp, err := ctx.Client.Get(ctx.Urls.Base + ctx.Urls.GetDomains)
	if err != nil {
		fmt.Printf("Error getting domains: %s", err.Error())
		return nil, errors.New("Unable to get")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("failed to fetch data: %d %s", resp.StatusCode, resp.Status)
		return nil, errors.New("HTTP error")
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil, errors.New("Unable to query")
	}

	li := doc.Find("li font").Text()
	if li != "" {
		err = errors.New(li)
	}

	m := map[string]string{}
	d := doc.Find("table").Eq(5)
	d.Find("tr td font").Each(func(i int, s *goquery.Selection) {
		b := s.Find("b")
		domain := b.Text()
		if strings.Contains(domain, ".") {
			href, _ := b.Parent().Parent().Find("a").Eq(0).Attr("href")
			domain_id := strings.Split(href, "=")[1]
			m[domain] = domain_id
			m[domain_id] = domain
		}
	})
	//	b, _ := io.ReadAll(resp.Body)

	return m, err
}

func (ctx *FreeDNS) CreateDomain(domain string) error {
	resp, err := ctx.Client.Get(ctx.Urls.Base + strings.Replace(ctx.Urls.CreateDomain, "{DOMAIN}", domain, -1))
	if err != nil {
		fmt.Printf("Error creating domain %s: %s", domain, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("failed to fetch data for domain %s: %d %s", domain, resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
		return errors.New("Unable to query")
	}

	li := doc.Find("li font").Text()
	if li != "" {
		err = errors.New(li)
	}

	return err
}

func (ctx *FreeDNS) DeleteDomain(domain_id string) error {
	resp, err := ctx.Client.Get(ctx.Urls.Base + strings.Replace(ctx.Urls.DeleteDomain, "{DOMAIN_ID}", domain_id, -1))
	if err != nil {
		fmt.Printf("Error deleting domain ID %s: %s", domain_id, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("failed to fetch data for domain ID %s: %d %s", domain_id, resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
		return errors.New("Unable to query")
	}

	li := doc.Find("li font").Text()
	if li != "" {
		err = errors.New(li)
	}

	return err
}

type Record struct {
	Id    string
	Name  string
	Type  string
	Value string
}

func (ctx *FreeDNS) GetRecords(domain_id string) (map[string]Record, error) {
	resp, err := ctx.Client.Get(ctx.Urls.Base + strings.Replace(ctx.Urls.GetRecords, "{DOMAIN_ID}", domain_id, -1))
	if err != nil {
		fmt.Printf("Error getting records for domain %s: %s", domain_id, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("failed to fetch data records for domain_id %s: %d %s", domain_id, resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
		return nil, errors.New("Unable to query")
	}

	li := doc.Find("li font").Text()
	if li != "" {
		err = errors.New(li)
	}

	m := map[string]Record{}
	d := doc.Find("form table tr")
	// fmt.Printf("d: %+v\n", d.Text())

	d.Each(func(i int, s *goquery.Selection) {
		a := s.Find("td a")
		domain := a.Text()
		if strings.Contains(domain, ".") {
			href, _ := a.Attr("href")
			record_id := strings.Split(href, "=")[1]

			t := a.Parent().Next().Text()
			v := a.Parent().Next().Next().Text()

			m[record_id] = Record{Id: record_id, Name: domain, Type: t, Value: v}
		}
	})

	return m, err
}

func (ctx *FreeDNS) UpdateRecord(domain_id string, record_id string, name string, t string, value string, ttl string) error {
	formData := url.Values{
		"domain_id": []string{domain_id},
		"subdomain": []string{name},
		"type":      []string{t},
		"address":   []string{value},
		"ttlalias":  []string{ttl},
	}
	if record_id != "" {
		formData["data_id"] = []string{record_id}
	}
	resp, err := ctx.Client.PostForm(ctx.Urls.Base+ctx.Urls.UpdateRecord, formData)
	if err != nil {
		fmt.Printf("Error creating record %s for dmain %s: %s", name, domain_id, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("failed to fetch data for domain ID %s record %s: %d %s", domain_id, name, resp.StatusCode, resp.Status)
	}
	// b, _ := io.ReadAll(resp.Body)
	// fmt.Printf("resp: %+v\n", string(b))

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
		return err
	}

	li := doc.Find("li font").Text()
	if li != "" {
		err = errors.New(li)
	}
	return err
}

func (ctx *FreeDNS) CreateRecord(domain_id string, name string, t string, value string, ttl string) error {
	return ctx.UpdateRecord(domain_id, "", name, t, value, ttl)
}

func (ctx *FreeDNS) FindRecordIds(m map[string]Record, name string) (ids []string, ok bool) {
	ok = false
	for k, v := range m {
		if v.Name == name {
			ids = append(ids, k)
			ok = true
		}
	}
	sort.Strings(ids)
	return ids, ok
}

func (ctx *FreeDNS) DeleteRecord(record_id string) error {
	formData := url.Values{
		"data_id": []string{record_id},
	}
	resp, err := ctx.Client.PostForm(ctx.Urls.Base+ctx.Urls.DeleteRecord, formData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(string(resp.StatusCode) + ":" + resp.Status)
	}
	// b, _ := io.ReadAll(resp.Body)
	// fmt.Printf("resp: %+v\n", string(b))

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
		return err
	}

	li := doc.Find("li font").Text()
	if li != "" {
		err = errors.New(li)
	}
	return err
}
