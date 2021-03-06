package postmortems

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	guuid "github.com/google/uuid"
)

var re = regexp.MustCompile(`^\[(.+?)\]\((.+?)\)\. (.+)$`)

var defaultBody = `---

uuid: "{{ .UUID }}"
url: ""
start_time: ""
end_time: ""
categories:
- postmortem
company: ""
product: ""

---

{{ .Data }}
`

// ExtractPostmortems reads the collection of postmortems
// and extracts each postmortem to a separate file.
func ExtractPostmortems(dir string) error {
	posts, err := ValidateDir(dir)
	if err != nil {
		return err
	}

	file, err := os.Open("./tmp/posts.md")
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		body := defaultBody

		// Generate a random string to set as UUID.
		id := guuid.New()
		pm := &Postmortem{UUID: id.String(), Description: scanner.Text()}

		if re.Match(scanner.Bytes()) {
			body = `---

uuid: "{{ .UUID }}"
url: "{{ .URL }}"
start_time: ""
end_time: ""
categories:
- postmortem
company: "{{ .Company }}"
product: ""

---

{{ .Description }}
`
			matches := re.FindStringSubmatch(scanner.Text())
			pm = &Postmortem{UUID: id.String(), URL: matches[2], Company: matches[1], Description: matches[3]}
		}

		// See if there is an existing one.
		for _, existing := range posts {
			if existing.URL == pm.URL {
				pm.UUID = existing.UUID
			}
		}

		err = pm.Save(body, dir)
		if err != nil {
			return fmt.Errorf("error saving postmortem file: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Save takes the in-memory representation of the postmortem file and stores it
// in a file.
func (pm *Postmortem) Save(body, dir string) error {
	var data bytes.Buffer

	fm := template.Must(template.New("newPostmortemTemplate").Parse(body))
	if err := fm.Execute(&data, pm); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	// Write postmortem data from memory to file.
	err := ioutil.WriteFile(filepath.Join(dir, pm.UUID+".md"), data.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}
