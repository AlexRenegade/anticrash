package main

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
)

func contains(a []string, b string) bool {
	for _, s := range a {
		if s == b {
			return true
		}
	}

	return false
}

func checkVideo(url string) bool {
	command := exec.Cmd{
		Path: ffProbePath,
		Args: []string{ffProbePath, "-v", "error", "-show_entries", "frame=pix_fmt,width,height", "-select_streams", "v", "-of", "csv=p=0", url},
	}

	stdout, _ := command.StdoutPipe()
	stderr, _ := command.StderrPipe()

	command.Start()

	scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
	lastLine := ""

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// An invalid block label doesn't symbolise a crash and should be ignored
		if strings.Contains(line, "invalid block label") {
			continue
		}
		// If a video file is missing a moov atom it cannot be probed, which causes a false positive
		if strings.Contains(line, "moov atom not found") {
			return false
		}
		if line != lastLine && lastLine != "" {
			_ = command.Process.Kill()
			return true
		}
		lastLine = line
	}

	_ = command.Wait()
	return false
}

func getURLsFromMessage(content string, attachments []*discordgo.MessageAttachment) []string {
	urlsToCheck := make([]string, 0)

	if len(attachments) != 0 {
		for _, attachment := range attachments {
			if strings.HasSuffix(attachment.URL, ".gif") || strings.HasSuffix(attachment.URL, ".mp4") {
				urlsToCheck = append(urlsToCheck, attachment.URL)
			}
		}
	}

	urlMatches := urlRegex.FindAllString(content, -1)

	for _, match := range urlMatches {
		if strings.Contains(match, "gfycat.com") && !strings.Contains(match, "giant.gfycat.com") && !strings.Contains(match, "gfycat.com/ifr/") {
			parts := strings.Split(match, ".com")
			if len(parts) > 1 {
				gfyName := gfyNameRegex.FindStringSubmatch(parts[1])

				if len(gfyName) > 1 {
					gfyURLs := getURLsFromGfyName(gfyName[1])
					urlsToCheck = append(urlsToCheck, gfyURLs...)
				}
			}
		} else if !strings.HasSuffix(match, ".mp4") && !strings.HasSuffix(match, ".gif") {
			urlsFromMeta := getURLsFromMeta(match)

			urlsToCheck = append(urlsToCheck, urlsFromMeta...)
		} else {
			urlsToCheck = append(urlsToCheck, match)
		}
	}

	return urlsToCheck
}

func getURLsFromMeta(url string) []string {
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("User-Agent", "AntiCrash/"+version)

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil
	}

	contentType := response.Header.Get("Content-Type")

	if strings.Contains(contentType, "image/gif") || strings.HasPrefix(contentType, "video/mp4") {
		return []string{url}
	}

	if !strings.Contains(contentType, "text/html") {
		return nil
	}

	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromReader(response.Body)
	content := make([]string, 0)

	if err != nil {
		return nil
	}

	content = append(content, getContentFromMeta(doc, `meta[property^="og:video"]`, "content")...)
	content = append(content, getContentFromMeta(doc, `meta[name^="twitter:player"]`, "content")...)
	content = append(content, getContentFromMeta(doc, `meta[property^="og:image"]`, "content")...)
	content = append(content, getContentFromMeta(doc, `meta[name^="twitter:image"]`, "content")...)

	doc.Find(`script[type="application/ld+json"]`).Each(func(_ int, selection *goquery.Selection) {
		urlMatches := urlRegex.FindAllString(selection.Text(), -1)

		content = append(content, urlMatches...)
	})

	urlsToReturn := make([]string, 0)

	for _, u := range content {
		if strings.HasSuffix(u, ".mp4") || strings.HasSuffix(u, ".gif") {
			if !contains(urlsToReturn, u) {
				urlsToReturn = append(urlsToReturn, u)
			}
		}
	}

	return urlsToReturn
}

func getContentFromMeta(doc *goquery.Document, selector string, attribute string) []string {
	elements := make([]string, 0)

	doc.Find(selector).Each(func(_ int, selection *goquery.Selection) {
		value, exists := selection.Attr(attribute)

		if !exists {
			return
		}

		elements = append(elements, value)
	})

	return elements
}

func getURLsFromGfyName(gfyName string) []string {
	url := "https://api.gfycat.com/v1/gfycats/" + gfyName
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("User-Agent", "AntiCrash/"+version)

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil
	}

	defer response.Body.Close()

	gfyResponse := struct {
		GfyItem struct {
			GifURL string `json:"gifUrl"`
			Mp4URL string `json:"mp4Url"`
		} `json:"gfyItem"`
	}{}
	err = json.Unmarshal(body, &gfyResponse)

	if err != nil {
		return nil
	}

	return []string{gfyResponse.GfyItem.Mp4URL, gfyResponse.GfyItem.GifURL}
}
