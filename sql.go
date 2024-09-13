package main

import (
	"errors"
	"math/rand/v2"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type state struct {
	Type   string
	attrib map[string]string
}

func randFrom[T any](arr []T) T {
	return arr[rand.IntN(len(arr))]
}

func hasReqiredPrevs(a []sqlAttrib, isPrev map[string]bool) bool {
	for _, attrib := range a {
		// attrib reqired and there it has no prev
		if notNull(attrib) && !isPrev[removeUnused(attrib.Ref())] {
			return false
		}
	}
	return true
}

func getPrecidance(data map[string]Table) ([]string, error) {
	nexts := make([]string, 0, len(data))
	for key := range data {
		nexts = append(nexts, key)
	}
	slices.Sort(nexts)
	parsed := make([]string, 0, len(data))
	isPrev := map[string]bool{"": true}
	for len(nexts) != 0 {
		writer := 0
		for _, e := range nexts {
			if hasReqiredPrevs(data[e], isPrev) {
				isPrev[e] = true
				parsed = append(parsed, e)
				// println("taken: " + e)
			} else {
				// println("reserved: " + e)
				nexts[writer] = e
				writer++
			}
		}
		// Error if nothing has been deleted
		if len(nexts) == writer {
			return parsed, errors.New("Cyclic reference detected in: " + strings.Join(nexts, ", "))
		}
		nexts = nexts[:writer]
	}
	return parsed, nil

}

func databaseFrom(data map[string]Table) string {
	precidence, err := getPrecidance(data)
	out := sqlNote(err)
	for _, name := range precidence {
		table := data[name]
		statments := []string{}
		for _, e := range table {
			statment := e.Key + " " + e.Type
			for _, e := range e.Meta {
				statment += " " + e
			}
			statment += ","
			if e.Discription != "" {
				statment += " -- " + e.Discription
			}
			statments = append(statments, statment)
		}

		primarys := []string{}
		for _, e := range table {
			if e.IsPrimary {
				primarys = append(primarys, e.Key)
			}
		}
		rules := []string{"PRIMARY KEY (" + strings.Join(primarys, ", ") + ")"}
		for _, e := range table {
			if e.Reference != "" {
				rules = append(rules, "FOREIGN KEY "+e.Key+" REFERENCES "+e.Reference)
			}
		}
		indent(statments, 1)
		indent(rules, 1)

		//TODO should i add `if not exists`
		out += "\n\nCREATE TABLE " + name + " (\n" + strings.Join(statments, "\n") + ",\n\n" + strings.Join(rules, ",\n") + "\n) ENGINE=innodb;"
	}
	return out
}

func sqlNote(e error) string {
	if e != nil {
		return "-- " + strings.ReplaceAll(e.Error(), "\n", "\n-- ") + "\n-- ~ The parsed data is shown below:\n\n"
	}
	return ""
}
func randPrice(max float64) string {
	return strconv.FormatFloat(rand.Float64()*max, 'f', 2, 64)
}

var names = []string{"john", "green", "frog", "max", "fumpt"}

var vals = map[string]func(state) string{
	"": func(s state) string {
		return "gibberish"
	},
	"size": func(s state) string {
		return randFrom([]string{
			"XS", "S", "M", "L", "XL",
		})
	},
	"price": func(s state) string { return randPrice(24.0) },
	"discount": func(s state) string {
		price, _ := strconv.ParseFloat(s.attrib["price"], 64)
		return randPrice(price)
	},
	"name": func(s state) string {
		return names[rand.IntN(len(names))]
	},
	"date": func(s state) string {
		return randTime("2006-01-02", 356*20)
	},
	// "phone":
	// "email":
	"description": func(s state) string {
		return randFrom([]string{
			"Lorem ipsum dolor sit amet",
			"consectetuer adipiscing elit",
			"Maecenas porttitor congue massa",
			"Fusce posuere",
			"magna sed pulvinar ultricies",
			"purus lectus malesuada libero",
			"sit amet commodo magna eros quis urna",
			"Nunc viverra imperdiet enim",
			"Fusce est",
			"Vivamus a tellus",
			"Pellentesque habitant morbi tristique senectus et",
			"netus et malesuada fames ac turpis egestas",
			"Proin pharetra nonummy pede",
		})
	},
}

func defaultGen(s state) string {
	word := regexp.MustCompile("[A-Za-z]+")
	typeMain := strings.ToLower(string(word.Find([]byte(s.Type))))
	if isText[typeMain] {
		return vals["description"](s) + typeMain
	}
	switch typeMain {
	case "int":
		return strconv.Itoa(rand.IntN(40))
	}
	return "[TODO]"
}

var isText = map[string]bool{
	"text": true,
	"blob": true,
	// "char":    true,
	"varchar": true,
}

func typeWrap(ty string, in string) string {
	if !regexp.MustCompile(`^[0-9\.]+$`).MatchString(in) || isText[strings.ToLower(ty)] {
		return strconv.Quote(in)
	}
	return in
}

func (t Table) getFeild(s string) sqlAttrib {
	for _, e := range t {
		if e.Key == s {
			return e
		}
	}
	return sqlAttrib{}
}

func insertionFrom(data map[string]Table, iter func(string) int) string {
	precidence, err := getPrecidance(data)
	note := sqlNote(err)
	insert := []string{}
	tableKeys := map[string][]map[string]string{}
	for _, name := range precidence {
		table := data[name]
		var feilds []string
		autoFeilds := []string{}
		for _, e := range table {
			if !e.autoIncimented() {
				feilds = append(feilds, e.Key)
			} else {
				autoFeilds = append(autoFeilds, e.Key)
			}
		}
		statments := []string{}
		keys := []map[string]string{}
		end := iter(name) + 1
		for i := 1; i != end; i++ {
			row := map[string]string{}
			statment := []string{}
			for _, e := range feilds {
				maxMatches := 0
				var generate = defaultGen
				words := split(e)
				for key, fun := range vals {
					funcWords := split(key)
					pos := matches(words, funcWords)
					if pos > maxMatches {
						maxMatches = pos
						generate = fun
					}
				}
				ty := data[name].getFeild(e).Type
				value := typeWrap(ty, generate(state{ty, row}))
				statment = append(statment, value)
				row[e] = value
			}
			for _, e := range autoFeilds {
				row[e] = strconv.Itoa(i)
			}
			keys = append(keys, row)
			// data, _ := json.Marshal(row)
			// statments = append(statments, "("+strings.Join(statment, ", ")+") -- "+string(data))
			statments = append(statments, "("+strings.Join(statment, ", ")+")")

		}
		tableKeys[name] = keys
		insert = append(insert, "INSERT INTO "+name+"\n("+strings.Join(feilds, ", ")+")\nVALUES\n"+strings.Join(statments, ",\n")+";")
	}
	return note + strings.Join(insert, "\n\n")
}

func handleEmptySql(s string) string {
	if strings.TrimSpace(s) != "" {
		return s
	}
	return "-- no data found"
}
