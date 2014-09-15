package cobe

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func tmpCopy(src string) (string, error) {
	sf, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer sf.Close()

	df, err := ioutil.TempFile("", "tests")
	if err != nil {
		return "", err
	}
	defer df.Close()

	io.Copy(df, sf)
	return df.Name(), nil
}

func TestInit(t *testing.T) {
	tmp, err := ioutil.TempFile("", "tests")
	if err != nil {
		t.Fatal(err)
	}
	path := tmp.Name()
	defer os.Remove(path)

	err = initGraph(path, defaultGraphOptions)
	if err != nil {
		t.Error(err)
	}
}

func TestInfo(t *testing.T) {
	filename, err := tmpCopy("data/pg11.brain")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(filename)

	g, err := openGraph(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer g.close()

	order, _ := g.getInfoString("order")
	if order != "3" {
		t.Errorf("Order: expected 3, was %s", order)
	}

	text, err := g.getInfoString("missing")
	if text != "" || err == nil {
		t.Error("Expected empty text and nil error")
	}

	err = g.setInfoString("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	text, err = g.getInfoString("foo")
	if text != "bar" || err != nil {
		t.Errorf("Expected bar, was %s", text)
	}
}

func TestAlice(t *testing.T) {
	filename, err := tmpCopy("data/pg11.brain")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(filename)

	g, err := openGraph(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer g.close()

	order, _ := g.getInfoString("order")
	if order != "3" {
		t.Errorf("Order: expected 3, was %s", order)
	}

	token, err := g.getTokenID("Alice")
	if err != nil {
		t.Error(err)
	}

	if token != 18 {
		t.Errorf("Token[Alice]: expected 18, was %d", token)
	}

	token = g.getOrCreateToken("Alice2")
	if token != 3428 {
		t.Errorf("Token[Alice2]: expected 3428, was %d", token)
	}

	if g.stemmer == nil {
		t.Error("Expected a non-nil stemmer")
	}
}

func TestKnownTokenIds(t *testing.T) {
	filename, err := tmpCopy("data/pg11.brain")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(filename)

	g, err := openGraph(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer g.close()

	// "." and "Alice" are found, "robot" not.
	knownIds := g.getKnownTokenIds([]string{".", "Alice", "robot"})
	if knownIds[0] != 11 {
		t.Errorf("Expected tokenId(\".\") == 11, was %d", knownIds[0])
	}

	if knownIds[1] != 18 {
		t.Errorf("Expected tokenId(\"Alice\") == 18, was %d", knownIds[1])
	}

	if len(knownIds) != 2 {
		t.Errorf("Expected 2 known tokenIds, was %d", len(knownIds))
	}

	// filter out non-words
	knownIds = g.filterWordTokenIds(knownIds)
	if knownIds[0] != 18 {
		t.Errorf("Expected tokenId(\"Alice\") == 18, was %d", knownIds[0])
	}

	if len(knownIds) != 1 {
		t.Errorf("Expected 1 known tokenId, was %d", len(knownIds))
	}
}

func TestGetTextByNode(t *testing.T) {
	filename, err := tmpCopy("data/pg11.brain")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(filename)

	g, err := openGraph(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer g.close()

	word, hasSpace, err := g.getTextByNodes(19, 20)
	if err != nil {
		t.Fatal(err)
	}

	if word != "." || hasSpace != true {
		t.Errorf("Expected . & true, got %s & %t", word, hasSpace)
	}
}
