package main

import (
	"io"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

func parse(r io.Reader) ([]*EmojiGroup, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse HTML")
	}

	contentDiv := findContentDiv(doc)
	if contentDiv == nil {
		return nil, errors.New(`did not find <div id="content">`)
	}

	var groups []*EmojiGroup
	for child := contentDiv.FirstChild; child != nil; child = child.NextSibling {
		if !isEmojiList(child) {
			continue
		}

		id, ok := getAttr(child, "id")
		if !ok {
			return nil, errors.Errorf("found emoji list without id attribute")
		}

		group, err := parseEmojiList(child)
		if err != nil {
			return nil, errors.Wrapf(err, "emoji list id=%q", id)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

func findContentDiv(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "id" && attr.Val == "content" {
				return n
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		content := findContentDiv(c)
		if content != nil {
			return content
		}
	}

	return nil
}

func isEmojiList(n *html.Node) bool {
	if n.Type != html.ElementNode || n.Data != "ul" {
		return false
	}

	return hasClass(n, "emojis")
}

func parseEmojiList(n *html.Node) (*EmojiGroup, error) {
	category, err := getEmojiListCategory(n)
	if err != nil {
		return nil, err
	}

	group := &EmojiGroup{Name: category}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode || child.Data != "li" {
			continue
		}

		emoji, err := parseEmojiListItem(child)
		if err != nil {
			return nil, err
		}

		group.Emojis = append(group.Emojis, emoji)
	}

	sort.Slice(group.Emojis, func(i, j int) bool {
		return group.Emojis[i].Name < group.Emojis[j].Name
	})

	return group, nil
}

func getEmojiListCategory(n *html.Node) (string, error) {
	for prev := n.PrevSibling; prev != nil; prev = prev.PrevSibling {
		if prev.Type != html.ElementNode || prev.Data != "h2" {
			continue
		}

		return prev.FirstChild.Data, nil
	}

	return "", errors.New("did not find h2 before emoji list")
}

func parseEmojiListItem(n *html.Node) (*Emoji, error) {
	if n.Type != html.ElementNode || n.Data != "li" {
		return nil, errors.New("expected emoji to be in an <li> element")
	}

	var div *html.Node
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}

		div = child
		break
	}

	if div == nil {
		return nil, errors.New("expected div element")
	}

	return parseEmojiDiv(div)
}

func parseEmojiDiv(n *html.Node) (*Emoji, error) {
	if n.Data != "div" {
		return nil, errors.New("expected emoji to contain a <div> element")
	}

	var emoji *Emoji
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode || c.Data != "span" {
			continue
		}

		if !hasClass(c, "name") {
			continue
		}

		emoji = &Emoji{Name: c.FirstChild.Data}
	}

	if emoji == nil {
		return nil, errors.New("did not find emoji span")
	}

	return emoji, nil
}

func getAttr(n *html.Node, name string) (string, bool) {
	for _, attr := range n.Attr {
		if attr.Key == name {
			return attr.Val, true
		}
	}

	return "", false
}

func hasClass(n *html.Node, wanted string) bool {
	classes, ok := getAttr(n, "class")
	if !ok {
		return false
	}

	for _, actual := range strings.Split(classes, " ") {
		if actual == wanted {
			return true
		}
	}

	return false
}
