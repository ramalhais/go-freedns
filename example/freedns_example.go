package main

import (
	freedns "github.com/ramalhais/go-freedns"
	log "github.com/sirupsen/logrus"
)

func main() {
	var err error

	log.Info("Testing go-freedns.")
	log.Info("Usage:")
	log.Infof("\tAUTH_LOGIN=you@example.com AUTH_PASSWORD=secret go run freedns_test.go\n")

	ctx, err := freedns.NewFreeDNS()
	if err != nil {
		log.Fatalf("Unable to create FreeDNS object: %s\n", err)
	}
	log.Debugf("Context: %+v\n", ctx)

	domains, _ := ctx.GetDomains()
	log.Debugf("Domains: %+v\n", domains)

	domain := "kube.ml"
	err = ctx.CreateDomain(domain)
	if err != nil {
		log.Errorf("Unable to create domain %s: %s\n", domain, err.Error())
	}

	domains, _ = ctx.GetDomains()
	log.Debugf("Domains: %+v\n", domains)

	recordName := "xpto"
	err = ctx.CreateRecord(domains[domain], recordName, "A", "8.8.8.8", "300")
	if err != nil {
		log.Errorf("Unable to create record %s on domain %s: %s\n", recordName, domain, err.Error())
	}

	records, _ := ctx.GetRecords(domains[domain])
	mlRecordIds, _ := ctx.FindRecordIds(records, recordName+domain)
	log.Debugf("mlRecordIds: %+v\n", mlRecordIds)
	err = ctx.UpdateRecord(domains[domain], mlRecordIds[0], recordName, "A", "8.8.8.8", "300")
	if err != nil {
		log.Errorf("Unable to update record: %s\n", err.Error())
	}

	records, _ = ctx.GetRecords(domains[domain])
	log.Debugf("Records: %+v\n", records)

	err = ctx.DeleteDomain(domains[domain])
	domains, _ = ctx.GetDomains()
	log.Debugf("Domains: %+v\n", domains)
}
