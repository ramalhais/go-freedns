package main

import (
	freedns "github.com/ramalhais/go-freedns"
	"github.com/sirupsen/logrus"
)

func main() {
	var err error

	log := logrus.New()
	log.Level = logrus.DebugLevel
	log.Info("Testing go-freedns.")
	log.Info("Usage:")
	log.Infof("\tAUTH_LOGIN=you@example.com AUTH_PASSWORD=secret go run freedns_test.go\n")
	log.Info("or")
	log.Infof("\tAUTH_COOKIE_VALUE=VALUE_OF_COOKIE_NAMED_dns_cookie go run freedns_test.go\n")

	ctx, err := freedns.NewFreeDNS()
	if err != nil {
		log.Fatalf("Unable to create FreeDNS object: %s\n", err)
	}
	log.Debugf("Context: %+v\n", ctx)

	// Get all DNS domains in your account
	domains, _, _ := ctx.GetDomains()
	log.Debugf("Domains: %+v\n", domains)

	// Create DNS domain
	domain := "domain.example"
	err = ctx.CreateDomain(domain)
	if err != nil {
		log.Infof("Unable to create domain %s: %s\n", domain, err.Error())
	}

	// Get all DNS domains in your account
	domains, _, _ = ctx.GetDomains()
	log.Debugf("Domains: %+v\n", domains)

	// Create DNS record
	recordName := "xpto"
	err = ctx.CreateRecord(domains[domain], recordName, "A", "8.8.8.8", "300")
	if err != nil {
		log.Errorf("Unable to create record %s on domain %s: %s\n", recordName, domain, err.Error())
	}

	// Get all DNS records for domain
	records, _ := ctx.GetRecords(domains[domain])
	// Find DNS record by FQDN in retrieved records
	recordIds, _ := ctx.FindRecordIds(records, recordName+"."+domain)
	log.Debugf("recordIds: %+v\n", recordIds)
	// Update first DNS record found
	err = ctx.UpdateRecord(domains[domain], recordIds[0], recordName, "A", "8.8.8.8", "300")
	if err != nil {
		log.Errorf("Unable to update record: %s\n", err.Error())
	}

	// Get all DNS records for domain
	records, _ = ctx.GetRecords(domains[domain])
	log.Debugf("Records: %+v\n", records)

	// Delete first DNS record found
	err = ctx.DeleteRecord(recordIds[0])

	// Get all DNS records for domain
	records, _ = ctx.GetRecords(domains[domain])
	log.Debugf("Records: %+v\n", records)

	records, _ = ctx.GetRecords(domains[domain])
	log.Debugf("Details records: %+v\n", records)
	recordIds, _ = ctx.FindRecordIds(records, "xdns-demo2."+domain)
	log.Debugf("Details recordIds: %+v\n", recordIds)
	if len(recordIds)>0 {
		details, _ := ctx.GetRecordDetails(recordIds[0])
		log.Debugf("Details: %+v\n", details)
	}

	// Delete DNS domain
	err = ctx.DeleteDomain(domains[domain])
	domains, _, _ = ctx.GetDomains()
	log.Debugf("Domains: %+v\n", domains)
}
