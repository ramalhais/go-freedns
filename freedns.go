package freedns

import (
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"strconv"
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
	GetRecordDetails string `envconfig:"URLS_GET_RECORD_DETAILS"`
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
	if ctx.Auth.CookieValue == "" && (ctx.Auth.Login == "" || ctx.Auth.Password == "") {
		return "", errors.New("Auth not found in configuration")
	}
	formData := url.Values{
		"username": []string{ctx.Auth.Login},
		"password": []string{ctx.Auth.Password},
		"action":   []string{"auth"},
	}

	if ctx.Auth.CookieValue != "" {
		cookie := &http.Cookie{
			Name:   ctx.Auth.CookieName,
			Value:  ctx.Auth.CookieValue,
			// MaxAge: 300,
		}
		url, _ := url.Parse(ctx.Urls.Base+ctx.Urls.Login)
		ctx.Client.Jar.SetCookies(url, []*http.Cookie{cookie})
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
			GetRecordDetails:   "/subdomain/edit.php?data_id={RECORD_ID}",
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
		log.Printf("ConfigFile: %s\n", err)
	}
	err = ctx.ConfigEnv()
	if err != nil {
		log.Printf("Error: ConfigEnv: %s", err)
	}

	jar, _ := cookiejar.New(nil)
	ctx.Client = &http.Client{
		Jar: jar,
	}

	_, err = ctx.Authenticate()

	return ctx, err
}

func (ctx *FreeDNS) GetDomains() (map[string]string, map[string]string, error) {
	resp, err := ctx.Client.Get(ctx.Urls.Base + ctx.Urls.GetDomains)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = errors.New("HTTP error " + strconv.Itoa(resp.StatusCode) + ": " + resp.Status)
		return nil, nil, err
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	li := doc.Find("li font").Text()
	if li != "" {
		err = errors.New(li)
	}

	mid := map[string]string{}
	mname := map[string]string{}
	d := doc.Find("table").Eq(6)
	d.Find("tr td font").Each(func(i int, s *goquery.Selection) {
		b := s.Find("b")
		domain := b.Text()
		if strings.Contains(domain, ".") {
			href, _ := b.Parent().Parent().Find("a").Eq(0).Attr("href")
			domain_id := strings.Split(href, "=")[1]
			mname[domain] = domain_id
			mid[domain_id] = domain
		}
	})
	//	b, _ := io.ReadAll(resp.Body)

	return mname, mid, err
}

func (ctx *FreeDNS) CreateDomain(domain string) error {
	resp, err := ctx.Client.Get(ctx.Urls.Base + strings.Replace(ctx.Urls.CreateDomain, "{DOMAIN}", domain, -1))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = errors.New("HTTP error " + strconv.Itoa(resp.StatusCode) + ": " + resp.Status)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
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
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = errors.New("HTTP error " + strconv.Itoa(resp.StatusCode) + ": " + resp.Status)
		return err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
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

type RecordDetails struct {
	Id    string
	Fqdn  string
	Type  string
	Host  string
	DomainId string
	Domain string
	Value string
	Ttl	  string
	Wildcard string
}

func (ctx *FreeDNS) GetRecordDetails(record_id string) (*RecordDetails, error) {
	record := RecordDetails{}

	resp, err := ctx.Client.Get(ctx.Urls.Base + strings.Replace(ctx.Urls.GetRecordDetails, "{RECORD_ID}", record_id, -1))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = errors.New("HTTP error " + strconv.Itoa(resp.StatusCode) + ": " + resp.Status)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	li := doc.Find("li font").Text()
	if li != "" {
		err = errors.New(li)
	}

	d := doc.Find("form table tr")
	// log.Printf("d: %+v\n", d.Text())

	record.Id = record_id
	record.Fqdn = strings.Replace(d.Eq(0).Text(), "Editing ", "", -1)
	record.Type = d.Eq(1).Find("td").Eq(1).Find("option[selected]").AttrOr("value", "")
	record.Host = d.Eq(2).Find("td").Eq(1).Find("input").AttrOr("value", "")
	record.DomainId = d.Eq(3).Find("td").Eq(1).Find("option[selected]").AttrOr("value", "")
	record.Domain = strings.Split(d.Eq(3).Find("td").Eq(1).Find("option[selected]").Text(), " ")[0]
	record.Value = d.Eq(4).Find("td").Eq(1).Find("input").AttrOr("value", "")
	record.Ttl = d.Eq(5).Find("td").Eq(1).Find("input").AttrOr("value", "")
	record.Wildcard = d.Eq(6).Find("td").Eq(1).Find("input.checked").AttrOr("value", "0")

	return &record, nil
}

func (ctx *FreeDNS) GetRecords(domain_id string) (map[string]Record, error) {
	resp, err := ctx.Client.Get(ctx.Urls.Base + strings.Replace(ctx.Urls.GetRecords, "{DOMAIN_ID}", domain_id, -1))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = errors.New("HTTP error " + strconv.Itoa(resp.StatusCode) + ": " + resp.Status)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	li := doc.Find("li font").Text()
	if li != "" {
		err = errors.New(li)
	}

	m := map[string]Record{}
	d := doc.Find("form table tr")
	// log.Printf("d: %+v\n", d.Text())

	d.Each(func(i int, s *goquery.Selection) {
		a := s.Find("td a")
		domain := a.Text()
		if strings.Contains(domain, ".") {
			href, _ := a.Attr("href")
			record_id := strings.Split(href, "=")[1]

			t := a.Parent().Next().Text()
			v := a.Parent().Next().Next().Text()

			if strings.HasSuffix(v, "...") {
				recordDetails, _ := ctx.GetRecordDetails(record_id)
				v = recordDetails.Value
			}
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
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = errors.New("HTTP error " + strconv.Itoa(resp.StatusCode) + ": " + resp.Status)
		return err
	}
	// b, _ := io.ReadAll(resp.Body)
	// log.Printf("resp: %+v\n", string(b))

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
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
		err = errors.New(strconv.Itoa(resp.StatusCode) + ":" + resp.Status)
		return err
	}
	// b, _ := io.ReadAll(resp.Body)
	// log.Printf("resp: %+v\n", string(b))

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	li := doc.Find("li font").Text()
	if li != "" {
		err = errors.New(li)
	}
	return err
}
