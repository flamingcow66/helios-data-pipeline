package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/flamingcow66/airtable"
)

type Directory struct {
	Students map[string]*Student
	Parents  map[string]*Parent
}

type Person struct {
	Email string
	Name  string
}

type Student struct {
	Person
	Class   string
	Grade   string
	Parents []*Parent
}

type Parent struct {
	Person
}

func main() {
	directoryPath := flag.String("directory", "", "path to directory CSV file")

	flag.Parse()

	l := slog.Default()

	if *directoryPath == "" {
		fatal(l, "please pass --directory")
	}

	dir, err := loadDirectory(l, *directoryPath)
	if err != nil {
		fatal(l, "failed to load directory", "error", err)
	}

	l.Info("loaded directory", "directory", dir)

	at, err := airtable.NewFromEnv()
	if err != nil {
		fatal(l, "failed to create airtable client", "error", err)
	}

	dirBase, err := at.GetBaseByName(context.Background(), "Directory")
	if err != nil {
		fatal(l, "failed to find Directory base in Airtable", "error", err)
	}

	parentsTable, err := dirBase.GetTableByName(context.Background(), "Parents")
	if err != nil {
		fatal(l, "failed to get Parents table", "error", err)
	}

	remoteParentsRecords, err := parentsTable.ListRecords(context.Background(), nil)
	if err != nil {
		fatal(l, "failed to fetch Parents", "error", err)
	}

	l.Info("tmp", "parents", remoteParentsRecords, "len", len(remoteParentsRecords))

	/*
		localParentRecords := dir.GetParentRecords()

		for _, record := range localParentRecords {
			_, err = parentsTable.AddRecords(&AirtableRecords{
				Records: []*AirtableRecord{record},
			})
			if err != nil {
				fatal(l, "failed to add Parents", "error", err)
			}
		}
	*/
}

func loadDirectory(l *slog.Logger, path string) (*Directory, error) {
	d := &Directory{
		Students: map[string]*Student{},
		Parents:  map[string]*Parent{},
	}

	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(fh)

	headers, err := r.Read()
	if err != nil {
		return nil, err
	}
	l.Info("headers", "headers", headers)

	iStudentFirstName := slices.Index(headers, "First Name")
	if iStudentFirstName == -1 {
		return nil, fmt.Errorf("'First Name' field missing")
	}

	iStudentLastName := slices.Index(headers, "Last Name")
	if iStudentLastName == -1 {
		return nil, fmt.Errorf("'Last Name' field missing")
	}

	iStudentClass := slices.Index(headers, "Class")
	if iStudentClass == -1 {
		return nil, fmt.Errorf("'Class' field missing")
	}

	iStudentGrade := slices.Index(headers, "Grade")
	if iStudentClass == -1 {
		return nil, fmt.Errorf("'Grade' field missing")
	}

	iParent1FirstName := slices.Index(headers, "Parent 1\nFirst")
	if iParent1FirstName == -1 {
		return nil, fmt.Errorf("'Parent 1\\nFirst' field missing")
	}

	iParent1LastName := slices.Index(headers, "Parent 1\nLast")
	if iParent1LastName == -1 {
		return nil, fmt.Errorf("'Parent 1\\nLast' field missing")
	}

	iParent1Email := slices.Index(headers, "Parent 1\nEmail")
	if iParent1Email == -1 {
		return nil, fmt.Errorf("'Parent 1\\nEmail' field missing")
	}

	iParent2FirstName := slices.Index(headers, "Parent 2\nFirst")
	if iParent2FirstName == -1 {
		return nil, fmt.Errorf("'Parent 2\\nFirst' field missing")
	}

	iParent2LastName := slices.Index(headers, "Parent 2\nLast")
	if iParent2LastName == -1 {
		return nil, fmt.Errorf("'Parent 2\\nLast' field missing")
	}

	iParent2Email := slices.Index(headers, "Parent 2 Email")
	if iParent2Email == -1 {
		return nil, fmt.Errorf("'Parent 2 Email' field missing")
	}

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		class := row[iStudentClass]
		if strings.Contains(class, "/") {
			class = ""
		}

		parents := []*Parent{}

		parent1 := d.AddParent(
			strings.ToLower(row[iParent1Email]),
			fmt.Sprintf("%s %s", row[iParent1FirstName], row[iParent1LastName]),
		)

		parents = append(parents, parent1)

		if row[iParent2Email] != "" {
			parent2 := d.AddParent(
				strings.ToLower(row[iParent2Email]),
				fmt.Sprintf("%s %s", row[iParent2FirstName], row[iParent2LastName]),
			)

			parents = append(parents, parent2)
		}

		studentEmail := fmt.Sprintf(
			"%s.%s@heliosschool.org",
			strings.ToLower(row[iStudentFirstName]),
			strings.ToLower(row[iStudentLastName]),
		)

		d.AddStudent(
			studentEmail,
			fmt.Sprintf("%s %s", row[iStudentFirstName], row[iStudentLastName]),
			class,
			row[iStudentGrade],
			parents,
		)
	}

	return d, nil
}

func fatal(l *slog.Logger, msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}

func (d *Directory) AddParent(email, name string) *Parent {
	p := d.Parents[email]
	if p != nil {
		return p
	}

	p = &Parent{
		Person: Person{
			Email: email,
			Name:  name,
		},
	}

	d.Parents[p.Email] = p

	return p
}

func (d *Directory) AddStudent(email, name, class, grade string, parents []*Parent) *Student {
	s := &Student{
		Person: Person{
			Email: email,
			Name:  name,
		},
		Class:   class,
		Grade:   grade,
		Parents: parents,
	}

	d.Students[email] = s

	return s
}

/*
func (d *Directory) GetParentRecords() []*AirtableRecord {
	records := []*AirtableRecord{}

	for _, parent := range d.Parents {
		records = append(records, &AirtableRecord{
			Fields: map[string]any{
				"Email": parent.Email,
				"Name":  parent.Name,
			},
		})
	}

	return records
}
*/

func (p Person) String() string {
	return fmt.Sprintf(
		"%s <%s>",
		p.Name,
		p.Email,
	)
}

func (s Student) String() string {
	return s.Person.String()
}

func (p Parent) String() string {
	return p.Person.String()
}
