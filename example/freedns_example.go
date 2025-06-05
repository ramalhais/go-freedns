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
	log.Infof("AUTH_LOGIN=you@example.com AUTH_PASSWORD=secret go run freedns_example.go")
	log.Info("or")
	log.Infof("AUTH_COOKIE_VALUE=VALUE_OF_COOKIE_NAMED_dns_cookie go run freedns_example.go")

	ctx, err := freedns.NewFreeDNS()
	if err != nil {
		log.Fatalf("Unable to create FreeDNS object: %s\n", err)
	}
	log.Debugf("Context: %+v", ctx)

	// Get all DNS domains in your account
	domains, _, _ := ctx.GetDomains()
	log.Debugf("Domains: %+v", domains)

	// Create DNS domain
	domain := "domain.example"
	err = ctx.CreateDomain(domain)
	if err != nil {
		log.Infof("Unable to create domain %s: %s", domain, err.Error())
	}

	// Get all DNS domains in your account
	domains, _, _ = ctx.GetDomains()
	log.Debugf("Domains: %+v", domains)

	// Create DNS record
	recordName := "xpto"
	err = ctx.CreateRecord(domains[domain], recordName, "A", "1.1.1.1", "300")
	if err != nil {
		log.Errorf("Unable to create record %s on domain %s: %s", recordName, domain, err.Error())
	}

	// Get all DNS records for domain
	records, _ := ctx.GetRecords(domains[domain])
	log.Debugf("Records: %+v", records)
	// Find DNS record by FQDN in retrieved records
	recordIds, _ := ctx.FindRecordIds(records, recordName+"."+domain)
	log.Debugf("recordIds: %+v", recordIds)
	// Update first DNS record found
	err = ctx.UpdateRecord(domains[domain], recordIds[0], recordName, "A", "8.8.8.8", "300")
	if err != nil {
		log.Errorf("Unable to update record: %s", err.Error())
	}

	// Get all DNS records for domain
	records, _ = ctx.GetRecords(domains[domain])
	log.Debugf("Records: %+v", records)

	// Delete first DNS record found
	err = ctx.DeleteRecord(recordIds[0])

	// Get all DNS records for domain
	records, _ = ctx.GetRecords(domains[domain])
	log.Debugf("Records: %+v", records)

	records, _ = ctx.GetRecords(domains[domain])
	log.Debugf("Details records: %+v", records)
	recordIds, _ = ctx.FindRecordIds(records, "xdns-demo2."+domain)
	log.Debugf("Details recordIds: %+v", recordIds)
	if len(recordIds)>0 {
		details, _ := ctx.GetRecordDetails(recordIds[0])
		log.Debugf("Details: %+v", details)
	}

	// Delete DNS domain
	err = ctx.DeleteDomain(domains[domain])
	domains, _, _ = ctx.GetDomains()
	log.Debugf("Domains: %+v", domains)
}
