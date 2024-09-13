package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func parseTable(doc *html.Node) map[string]Table {
	database := map[string]Table{}
	table := ""
	CapturedStartCells := 0
	for _, row := range goquery.NewDocumentFromNode(doc).Find("tr").Nodes {
		cells := goquery.NewDocumentFromNode(row).Find("td").Nodes
		if CapturedStartCells == 0 {
			text := getText(cells[0])
			if text != "" {
				table = text
			}
			CapturedStartCells = 1
			for _, arr := range cells[0].Attr {
				if arr.Key == "rowspan" {
					var err error
					CapturedStartCells, err = strconv.Atoi(arr.Val)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
			cells = cells[1:]
		}
		meta, ref, isPrimary, discription := fromDiscription(tryChild(cells, 2))
		database[table] = append(database[table], sqlAttrib{
			Key:         collapseWhiteSpace(tryChild(cells, 0)),
			Type:        collapseWhiteSpace(tryChild(cells, 1)),
			Meta:        meta,
			Reference:   ref,
			IsPrimary:   isPrimary,
			Discription: discription,
		})

		CapturedStartCells--
	}
	return database
}

func collapseWhiteSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func removeUnused(s string) string {
	return regexp.MustCompile(`(?i)(^[^A-Z0-9\(\)]*)|([^A-Z0-9;\(\)]*$)`).ReplaceAllString(s, "")
}

// Gets text from node as it would apper in html if tags were trimed
// all whitespace will be collapsed
func getText(node *html.Node) string {
	// spaces/tabs/newlines are often used as code formating in html
	// but are ignored in parsing
	return collapseWhiteSpace(goquery.NewDocumentFromNode(node).Text())
}

type sqlAttrib struct {
	Key         string
	Type        string
	Meta        []string
	Reference   string
	IsPrimary   bool
	Discription string
}

func (a sqlAttrib) Ref() string {
	return string(
		regexp.MustCompile(
			`(?i)[a-z]+`,
		).Find(
			[]byte(a.Reference),
		),
	)
}

func (a sqlAttrib) autoIncimented() bool {
	for _, e := range a.Meta {
		if e == "AUTO_INCRIMENT" {
			return true
		}
	}
	return false
}

type Table []sqlAttrib

func tryChild(children []*html.Node, index int) string {
	if uint(index) < uint(len(children)) {
		return getText(children[index])
	}
	return ""
}

func indent(arr []string, level int) {
	for i, e := range arr {
		arr[i] = strings.Repeat("\t", level) + removeUnused(e)
	}
}

func notNull(attrib sqlAttrib) bool {
	// print(attrib.Key + ": ")
	for _, e := range attrib.Meta {
		if e == "NOT NULL" {
			// println("not null")
			return true
		}
	}
	// println("nullable")
	return false
}

func fromDiscription(discription string) (meta []string, ref string, primary bool, modifiedDiscription string) {
	if isPrimary, _ := regexp.MatchString(`(?i)PRIMARY`, discription); isPrimary {
		primary = true
	}
	discription = regexp.MustCompile(`(?i)\s*(PRIMARY|KEY)\s*?`).ReplaceAllString(discription, " ")
	null := regexp.MustCompile(`(?i)\s*NOT.*?NULL\s*`)
	if null.Find([]byte(discription)) != nil {
		meta = append(meta, "NOT NULL")
		discription = null.ReplaceAllString(discription, " ")
	}

	incriment := regexp.MustCompile(`(?i)(AUTO.*)?INCR(I|E)MENT`)
	if incriment.Find([]byte(discription)) != nil {
		meta = append(meta, "AUTO_INCRIMENT")
		discription = incriment.ReplaceAllString(discription, " ")
	}

	referenceAtom := regexp.MustCompile(`(?i)REFERENCE.*?\)`)
	if atom := referenceAtom.Find([]byte(discription)); atom != nil {
		external := regexp.MustCompile(`(?i)[^\s]*\(.*?\)`)
		ref = collapseWhiteSpace(string(external.Find(atom)))
		discription = referenceAtom.ReplaceAllString(discription, " ")
	}
	modifiedDiscription = collapseWhiteSpace(removeUnused(discription))
	return
}

type filler = struct{}

func split(s string) map[string]filler {
	r := regexp.MustCompile("[A-Za-z][a-z]+")
	res := map[string]filler{}
	for _, e := range r.FindAllString(s, -1) {
		res[strings.ToLower(e)] = filler{}
	}
	return res
}

func matches(a, b map[string]filler) int {
	res := 0
	for i, _ := range a {
		if _, ok := b[i]; ok {
			res++
		}
	}
	return res
}
