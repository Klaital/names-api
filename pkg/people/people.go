package people

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
)

type NameFile struct {
	Names []Person `yaml:"names"`
}

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
	Name         string   `json:"name" yaml:"name"`
	Description  string   `json:"descr" yaml:"description"`
	Source       string   `json:"source" yaml:"source"`
	Gender       string   `json:"gender" yaml:"gender"`
	ReferenceUrl string   `json:"ref" yaml:"reference_url"`
	Nicknames    []string `json:"nicknames" yaml:"nicknames"`
	Tags         []string `json:"tags" yaml:"tags"`
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

// Update adds and removes tags and nicknames in
// addition to updating the Person's fields.
func (p *Person) Update(updatedData *Person, db *sqlx.DB) error {
	if p.Id == 0 {
		return errors.New("cannot update person without an ID")
	}

	// Update the base person record
	sqlStmt := db.Rebind(`UPDATE people SET name=:name, descr=:descr, source=:source, gender=:gender, ref=:ref WHERE id=:id`)
	_, err := db.NamedExec(sqlStmt, p.ToPersonRecord())
	if err != nil {
		return err
	}

	// Add and remove nicknames
	existingNames := make(map[string]bool, 0)
	for i := range p.Nicknames {
		existingNames[p.Nicknames[i]] = true
	}
	updatedNames := make(map[string]bool, 0)
	for i := range updatedData.Nicknames {
		updatedNames[updatedData.Nicknames[i]] = true
	}

	// Find new names to add
	sqlStmt = db.Rebind(`INSERT INTO person_names (person_id, name) VALUES (?, ?)`)
	for name := range updatedNames {
		if !existingNames[name] {
			_, err := db.Exec(sqlStmt, p.Id, name)
			if err != nil {
				return err
			}
		}
	}
	// Find missing names to remove
	existingTags := make(map[string]bool, 0)
	for i := range p.Tags {
		existingNames[p.Tags[i]] = true
	}
	updatedTags := make(map[string]bool, 0)
	for i := range updatedData.Tags {
		updatedNames[updatedData.Tags[i]] = true
	}
	// Find new tags to add
	sqlStmt = db.Rebind(`INSERT INTO person_tags (person_id, name) VALUES (?, ?)`)
	for t := range updatedTags {
		if !existingTags[t] {
			_, err := db.Exec(sqlStmt, p.Id, t)
			if err != nil {
				return err
			}
		}
	}
	sqlStmt = db.Rebind(`DELETE FROM person_tags WHERE person_id = ? AND tag = ?`)
	for t := range existingTags {
		if !updatedTags[t] {
			_, err := db.Exec(sqlStmt, p.Id, t)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
func (p *Person) Insert(db *sqlx.DB) error {
	// Insert the base person, fetch their new ID, then insert the tags and nicknames
	sqlStmt := db.Rebind(`INSERT INTO people (name, descr, source, gender, ref) VALUES (:name, :descr, :source, :gender, :ref)`)
	resp, err := db.NamedExec(sqlStmt, p.ToPersonRecord())
	if err != nil {
		return err
	}
	id, err := resp.LastInsertId()
	if err != nil {
		return err
	}
	if id == 0 {
		return errors.New("invalid row ID returned")
	}

	sqlStmt = db.Rebind(`INSERT INTO people_tags (person_id, tag) VALUES (:person_id, :tag)`)
	for i := range p.Tags {
		_, err = db.NamedExec(sqlStmt, PersonTag{
			PersonId: int(id),
			Tag:      p.Tags[i],
		})
		if err != nil {
			return err
		}
	}

	sqlStmt = db.Rebind(`INSERT INTO people_names (person_id, name) VALUES (:person_id, :name)`)
	for i := range p.Nicknames {
		_, err = db.NamedExec(sqlStmt, Nickname{
			PersonId: int(id),
			Name:     p.Nicknames[i],
		})
		if err != nil {
			return err
		}
	}

	return nil
}
