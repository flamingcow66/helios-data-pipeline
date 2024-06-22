package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
)

type Person struct {
	FirstName string
	LastName  string
	Email     string
}

type Student struct {
	Person
	Class string
	Grade string
}

func main() {
	directoryPath := flag.String("directory", "", "path to directory CSV file")

	flag.Parse()

	l := slog.Default()

	if *directoryPath == "" {
		fatal(l, "please pass --directory")
	}

	_, err := loadDirectory(l, *directoryPath)
	if err != nil {
		fatal(l, "failed to load directory", "error", err)
	}
}

func loadDirectory(l *slog.Logger, path string) (*string, error) {
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

		studentEmail := fmt.Sprintf(
			"%s.%s@heliosschool.org",
			strings.ToLower(row[iStudentFirstName]),
			strings.ToLower(row[iStudentLastName]),
		)

		student := &Student{
			Person: Person{
				FirstName: row[iStudentFirstName],
				LastName:  row[iStudentLastName],
				Email: studentEmail,
			},
			Class: class,
			Grade: row[iStudentGrade],
		}

		l.Info("row", "row", student)
	}

	return nil, nil
}

func fatal(l *slog.Logger, msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}
