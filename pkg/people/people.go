package people

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
)

// PersonRecord models the database row for the 'people' table
type PersonRecord struct {
	Id           int            `db:"id"`
	Name         string         `db:"name"`
	Description  sql.NullString `db:"descr"`
	Source       sql.NullString `db:"source"`
	Gender       sql.NullString `db:"gender"`
	ReferenceUrl sql.NullString `db:"ref"`
}

type Person struct {
	Id           int      `json:"-"`
	Name         string   `json:"name"`
	Description  string   `json:"descr"`
	Source       string   `json:"source"`
	Gender       string   `json:"gender"`
	ReferenceUrl string   `json:"ref"`
	Nicknames    []string `json:"nicknames"`
	Tags         []string `json:"tags"`
}
type Nickname struct {
	Id       int    `json:"-" db:"id"`
	PersonId int    `json:"-" db:"person_id"`
	Name     string `json:"name" db:"name"`
}
type PersonTag struct {
	Id       int    `json:"-" db:"id"`
	PersonId int    `json:"-" db:"person_id"`
	Tag      string `json:"tag" db:"tag"`
}

func (p *PersonRecord) ToPerson() Person {
	person := Person{
		Id:   p.Id,
		Name: p.Name,
	}
	if p.Description.Valid {
		person.Description = p.Description.String
	}
	if p.Source.Valid {
		person.Source = p.Source.String
	}
	if p.Gender.Valid {
		person.Gender = p.Gender.String
	}
	if p.ReferenceUrl.Valid {
		person.ReferenceUrl = p.ReferenceUrl.String
	}
	return person
}

func (p *Person) ToPersonRecord() PersonRecord {
	record := PersonRecord{
		Id:           p.Id,
		Name:         p.Name,
		Description:  sql.NullString{String: p.Description},
		Source:       sql.NullString{String: p.Source},
		Gender:       sql.NullString{String: p.Gender},
		ReferenceUrl: sql.NullString{String: p.ReferenceUrl},
	}
	record.Description.Valid = len(p.Description) > 0
	record.Source.Valid = len(p.Source) > 0
	record.Gender.Valid = len(p.Gender) > 0
	record.ReferenceUrl.Valid = len(p.ReferenceUrl) > 0
	return record
}

func LoadAllPeople(db *sqlx.DB) ([]Person, error) {
	records := make([]PersonRecord, 0)
	err := db.Select(&records, db.Rebind(`SELECT id, name, descr, source, gender, ref FROM people`))
	if err != nil {
		return nil, err
	}

	// Convert nullable DB records into useful structs
	people := make(map[int]Person, len(records))
	for i := range records {
		people[records[i].Id] = records[i].ToPerson()
	}

	// Load nicknames
	nicknames := make([]Nickname, 0)
	err = db.Select(&nicknames, db.Rebind(`SELECT id, person_id, name FROM people_names`))
	if err != nil {
		return nil, err
	}
	for _, nickname := range nicknames {
		thisPerson := people[nickname.PersonId]
		nicks := thisPerson.Nicknames
		if nicks == nil {
			nicks = []string{nickname.Name}
		} else {
			nicks = append(nicks, nickname.Name)
		}
		thisPerson.Nicknames = nicks
		people[thisPerson.Id] = thisPerson
	}

	// load tags
	tags := make([]PersonTag, 0)
	err = db.Select(&tags, db.Rebind(`SELECT id, person_id, tag FROM people_tags`))
	if err != nil {
		return nil, err
	}
	for _, personTag := range tags {
		thisPerson := people[personTag.PersonId]
		thisPersonTags := thisPerson.Tags
		if thisPersonTags == nil {
			thisPersonTags = []string{personTag.Tag}
		} else {
			thisPersonTags = append(thisPersonTags, personTag.Tag)
		}
		thisPerson.Tags = thisPersonTags
		people[thisPerson.Id] = thisPerson
	}

	// Reformat back to a flat array of people
	peopleArray := make([]Person, 0, len(people))
	for i := range people {
		peopleArray = append(peopleArray, people[i])
	}

	return peopleArray, nil
}
