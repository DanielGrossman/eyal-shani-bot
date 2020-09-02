package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/ghodss/yaml"
)

const (
	male     = "male"
	female   = "female"
	singular = "singular"
	plural   = "plural"
)

type vocabFields struct {
	PrefixA     []string `yaml:"previxA"`
	IngridientA []string `yaml:"ingridientA"`
	PrefixB     []string `yaml:"prefixB"`
	IngridientB []string `yaml:"ingridientB"`
	Adjective   []string `yaml:"adjective"`
	Verb        []string `yaml:"verb"`
}

type Raw struct {
	Vocab struct {
		Names map[string]map[string]vocabFields `yaml:"names"`
		Place []string                          `yaml:"place"`
	} `yaml:"vocab"`
}

type definition struct {
	gender string
	form   string
}

type grammar struct {
	prefixA     definition
	IngridientA definition
	AdjectiveA  definition
	Verb        definition
	PrefixB     definition
	IngridientB definition
	AdjectiveB  definition
}

func main() {
	c := getConfig()

	if c.ConsumerKey == "" || c.ConsumerSecret == "" || c.AccessToken == "" || c.AccessSecret == "" {
		log.Fatal("Consumer key/secret and Access token/secret required")
	}

	config := oauth1.NewConfig(c.ConsumerKey, c.ConsumerSecret)
	token := oauth1.NewToken(c.AccessToken, c.AccessSecret)

	// OAuth1 http.Client will automatically authorize Requests
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter Client
	client := twitter.NewClient(httpClient)

	rand.Seed(time.Now().Unix())
	v, err := getVocab()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	t, err := time.ParseDuration(c.Period)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	ticker := time.NewTicker(t * time.Second)
	for range ticker.C {
		dish := makeDish(v)

		tweet, _, err := client.Statuses.Update(dish, nil)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		log.Println(tweet.FullText)
	}

}

func makeDish(raw *Raw) string {
	g := getGrammar()
	r := raw.Vocab.Names
	prefixA := chooseRandom(r[g.prefixA.form][g.prefixA.gender].PrefixA)
	ingridientA := chooseRandom(r[g.IngridientA.form][g.IngridientA.gender].IngridientA)
	adjectiveA := chooseRandom(r[g.AdjectiveA.form][g.AdjectiveA.gender].Adjective)
	placeA := chooseRandom(raw.Vocab.Place)
	verb := chooseRandom(r[g.Verb.form][g.Verb.gender].Verb)
	prefixB := chooseRandomExclude(r[g.PrefixB.form][g.PrefixB.gender].PrefixB, prefixA)
	ingridientB := chooseRandomExclude(r[g.IngridientB.form][g.IngridientB.gender].IngridientB, ingridientA)
	adjectiveB := chooseRandomExclude(r[g.AdjectiveB.form][g.AdjectiveB.gender].Adjective, adjectiveA)
	placeB := chooseRandomExclude(raw.Vocab.Place, placeA)
	in := chooseRandom([]string{"ב", " על"})
	from := "מ"

	dish := fmt.Sprintf("%s %s %s %s %s %s %s %s %s",
		prefixA,
		ingridientA,
		wordOrEmpty(adjectiveA),
		wordOrEmpty(from+placeA),
		chooseRandom([]string{",", verb + " " + in}),
		prefixB,
		ingridientB,
		wordOrEmpty(adjectiveB),
		wordOrEmpty(from+placeB))
	return strings.Replace(dish, "  ", " ", -1)
}

func getGrammar() grammar {
	var g grammar

	g.IngridientA.gender = chooseRandomGender()
	g.IngridientA.form = chooseRandomForm()
	if g.IngridientA.form == plural {
		g.IngridientA.gender = male
	}
	g.prefixA.form = chooseRandomForm()
	g.prefixA.gender = chooseRandomGender()

	g.AdjectiveA.form = chooseRandomForm()
	g.AdjectiveA.gender = chooseRandomGender()

	if g.IngridientA.gender == g.prefixA.gender {
		g.AdjectiveA.gender = g.IngridientA.gender
	}

	if g.prefixA.form == g.IngridientA.form {
		g.AdjectiveA.form = g.prefixA.form
	}

	if g.IngridientA.form != g.prefixA.form && g.IngridientA.gender != g.prefixA.gender {
		r := chooseRandomSlice([][]string{{g.IngridientA.form, g.IngridientA.gender}, {g.prefixA.form, g.prefixA.gender}})
		g.AdjectiveA.form = r[0]
		g.AdjectiveA.gender = r[1]
	}
	g.Verb.form = g.prefixA.form
	g.Verb.gender = g.prefixA.gender

	g.IngridientB.gender = chooseRandomGender()
	g.IngridientB.form = chooseRandomForm()
	if g.IngridientB.form == plural {
		g.IngridientB.gender = male
	}
	g.PrefixB.form = chooseRandomForm()
	g.PrefixB.gender = chooseRandomGender()

	g.AdjectiveB.form = chooseRandomForm()
	g.AdjectiveB.gender = chooseRandomGender()

	if g.IngridientB.gender == g.PrefixB.gender {
		g.AdjectiveB.gender = g.IngridientB.gender
	}

	if g.PrefixB.form == g.IngridientB.form {
		g.AdjectiveB.form = g.PrefixB.form
	}

	if g.IngridientB.form != g.PrefixB.form && g.IngridientB.gender != g.PrefixB.gender {
		r := chooseRandomSlice([][]string{{g.IngridientB.form, g.IngridientB.gender}, {g.PrefixB.form, g.PrefixB.gender}})
		g.AdjectiveB.form = r[0]
		g.AdjectiveB.gender = r[1]
	}
	return g
}

func wordOrEmpty(s string) string {
	return chooseRandom([]string{"", s})
}

func chooseRandomGender() string {
	return chooseRandom([]string{male, female})
}

func chooseRandomForm() string {
	return chooseRandom([]string{singular, plural})
}

func chooseRandomExclude(s []string, x string) string {
	var result string
	for {
		r := chooseRandom(s)
		if r != x {
			result = r
			break
		}
	}
	return result
}

func chooseRandom(s []string) string {
	return s[rand.Intn(len(s))]
}

func chooseRandomSlice(s [][]string) []string {
	return s[rand.Intn(len(s))]
}

func getVocab() (*Raw, error) {
	vocabFile, err := os.Open("vocab.yaml")
	if err != nil {
		return nil, err
	}

	byteValue, err := ioutil.ReadAll(vocabFile)
	if err != nil {
		return nil, err
	}
	var v Raw
	err = yaml.Unmarshal(byteValue, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
