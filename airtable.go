package main

import (
	"fmt"
	"os"

	"github.com/mehanizm/airtable"
)

type AirtableRecord = airtable.Record
type AirtableRecords = airtable.Records

type Airtable struct {
	c *airtable.Client
}

func NewAirtableFromEnv() (*Airtable, error) {
	airtableToken := os.Getenv("AIRTABLE_TOKEN")
	if airtableToken == "" {
		return nil, fmt.Errorf("please set $AIRTABLE_TOKEN")
	}

	return &Airtable{
		c: airtable.NewClient(airtableToken),
	}, nil
}

func (at *Airtable) GetBaseID(name string) (string, error) {
	bases, err := at.c.GetBases().WithOffset("").Do()
	if err != nil {
		return "", fmt.Errorf("failed to get bases: %w", err)
	}

	for _, base := range bases.Bases {
		if base.Name == name {
			return base.ID, nil
		}
	}

	return "", fmt.Errorf("base not found: %s", name)
}

func (at *Airtable) GetTable(baseID, name string) (*airtable.Table, error) {
	return at.c.GetTable(baseID, name), nil
}
