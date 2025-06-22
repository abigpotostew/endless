package train

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/mb-14/gomarkov"
)

type GeneratedPage struct {
	Link        PageLink
	Content     string
	Links       []PageLink
	LastUpdated time.Time
	Author      string
}

func GeneratePage(seed int64, chain *gomarkov.Chain) (GeneratedPage, error) {
	prng := rand.New(rand.NewSource(seed))
	thisLink, err := createLinkFromSeed(seed, prng, chain)
	if err != nil {
		return GeneratedPage{}, err
	}
	content, err := createParagraph(prng, chain)
	if err != nil {
		return GeneratedPage{}, err
	}
	links, err := createLinks(prng, chain)
	if err != nil {
		return GeneratedPage{}, err
	}
	lastUpdated := generateRandomDate(prng)
	author := authors[prng.Intn(len(authors))]

	page := GeneratedPage{
		Link:        thisLink,
		Content:     content,
		Links:       links,
		LastUpdated: lastUpdated,
		Author:      author,
	}
	return page, nil
}

func createParagraph(prng *rand.Rand, chain *gomarkov.Chain) (string, error) {
	sentenceCount := prng.Intn(10) + 1
	paragraph := ""
	for i := 0; i < sentenceCount; i++ {
		sentence, err := GenerateStoryFromPrng(prng, chain)
		if err != nil {
			return "", err
		}
		paragraph += sentence + " "
	}
	return paragraph, nil
}

func createNewLink(prngOld *rand.Rand, chain *gomarkov.Chain) (PageLink, error) {
	seed := prngOld.Int63()
	prng := rand.New(rand.NewSource(seed))
	return createLinkFromSeed(seed, prng, chain)
}

func createLinkFromSeed(seed int64, prng *rand.Rand, chain *gomarkov.Chain) (PageLink, error) {
	title, err := GenerateStoryFromPrng(prng, chain)
	if err != nil {
		return PageLink{}, err
	}
	//make this url friendly.
	//replace spaces with dashes
	link := strings.TrimSpace(title)
	link = strings.ToLower(link)
	link = strings.ReplaceAll(link, " ", "-")
	link = strings.ReplaceAll(link, "\n", "-")
	//now remove any non-alphanumeric characters
	link = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
			return r
		}
		return -1
	}, link)
	//now remove any duplicate dashes
	link = strings.ReplaceAll(link, "--", "-")
	//now remove any leading or trailing dashes
	link = strings.Trim(link, "-")

	return PageLink{
		Url:   fmt.Sprintf("/post/%d-%s", seed, link),
		Title: title,
		Seed:  seed,
	}, nil
}

type PageLink struct {
	Url   string
	Title string
	Seed  int64
}

func createLinks(prng *rand.Rand, chain *gomarkov.Chain) ([]PageLink, error) {
	linkCount := prng.Intn(3) + 1
	links := []PageLink{}
	for i := 0; i < linkCount; i++ {
		link, err := createNewLink(prng, chain)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

// generateRandomDate creates a random date within the last 2 years
func generateRandomDate(prng *rand.Rand) time.Time {
	// Generate a random date within the last 2 years
	now := time.Now()
	twoYearsAgo := now.AddDate(-2, 0, 0)

	// Generate random seconds since two years ago
	secondsRange := int64(now.Sub(twoYearsAgo).Seconds())
	randomSeconds := prng.Int63n(secondsRange)

	// Add random seconds to the base date
	randomDate := twoYearsAgo.Add(time.Duration(randomSeconds) * time.Second)

	return randomDate
}

var authors = []string{
	"Arlo Mills",
	"Joe Goetz",
	"Billy Goetz",
	"Marybeth Trott",
	"Charlie Davis",
	"Diana White",
	"Ethan Young",
}
