package szkopul_client

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Arapak/sio-tool/sio_submissions"
	"github.com/Arapak/sio-tool/util"

	"github.com/PuerkitoBio/goquery"
)

func getSubmissionID(body []byte) (string, error) {
	reg := regexp.MustCompile(`<tr id="report(\d+?)row">`)
	tmp := reg.FindSubmatch(body)
	if len(tmp) < 2 {
		return "", errors.New("cannot find submission id")
	}
	return string(tmp[1]), nil
}

func findSubmission(body []byte, n int) ([][]byte, error) {
	reg := regexp.MustCompile(`<tr id="report\d+row">[\s\S]+?</tr>`)
	tmp := reg.FindAll(body, n)
	if tmp == nil {
		return nil, errors.New("cannot find any submission")
	}
	return tmp, nil
}

func getProblemNames(name string) (string, string) {
	reg := regexp.MustCompile(`([\s\S]+?) \((\S+?)\)`)
	tmp := reg.FindSubmatch([]byte(name))
	if len(tmp) < 3 {
		return name, ""
	}
	return string(tmp[1]), string(tmp[2])
}

func parseSubmission(body []byte) (ret sio_submissions.Submission, err error) {
	id, err := getSubmissionID(body)
	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(fmt.Sprintf("<table>%v</table>", string(body))))
	if err != nil {
		return
	}
	get := func(sel string) string {
		return strings.TrimSpace(doc.Find(sel).Text())
	}
	when := strings.TrimSpace(doc.Find("a").First().Text())
	combinedName := get(fmt.Sprintf("td#submission%v-problem-instance", id))
	name, shortName := getProblemNames(combinedName)
	points := sio_submissions.ToInt(get(fmt.Sprintf("td#submission%v-score", id)))
	status := strings.ToLower(get(fmt.Sprintf("td#submission%v-status", id)))
	end := true
	if strings.Contains(strings.ToLower(status), "oczekuje") || strings.Contains(strings.ToLower(status), "pending") {
		end = false
	}
	if strings.Contains(strings.ToLower(status), "ok") {
		status = fmt.Sprintf("${c-accepted}%v", status)
		if points == sio_submissions.Inf {
			end = false
		}
	} else if strings.Contains(strings.ToLower(status), "błąd") || strings.Contains(strings.ToLower(status), "failed") {
		status = fmt.Sprintf("${c-failed}%v", status)
		if points == sio_submissions.Inf && !strings.Contains(strings.ToLower(status), "kompilacji") && !strings.Contains(strings.ToLower(status), "compilation") {
			end = false
		}
	} else {
		status = fmt.Sprintf("${c-rejected}%v", status)
	}
	return sio_submissions.Submission{
		Id:        sio_submissions.ToInt(id),
		Name:      name,
		ShortName: shortName,
		Status:    status,
		Points:    points,
		When:      when,
		End:       end,
	}, nil
}

func getProblemName(body []byte) string {
	reg := regexp.MustCompile(`<div class="problem-title text-center content-row">\s+?<h1>([\s\S]+?)</h1>`)
	tmp := reg.FindSubmatch(body)
	if len(tmp) < 2 {
		return ""
	}
	return string(tmp[1])
}

func GetSubmissions(client *http.Client, URL string, n int) (submissions []sio_submissions.Submission, err error) {
	body, err := util.GetBody(client, URL)
	if err != nil {
		return
	}

	if _, err = findUsername(body); err != nil {
		return
	}

	submissionsBody, err := findSubmission(body, n)
	if err != nil {
		return
	}

	name, shortName := getProblemNames(getProblemName(body))

	for _, submissionBody := range submissionsBody {
		if submission, err := parseSubmission(submissionBody); err == nil {
			if submission.Name == "" && submission.ShortName == "" {
				submission.Name = name
				submission.ShortName = shortName
			}
			submissions = append(submissions, submission)
		}
	}

	if len(submissions) < 1 {
		return nil, errors.New("cannot find any submission")
	}

	return
}

func (c *SzkopulClient) WatchSubmission(info Info, n int, line bool) (submissions []sio_submissions.Submission, err error) {
	URL := info.MySubmissionURL(c.host)
	if err != nil {
		return
	}

	maxWidth := 0
	first := true
	for {
		st := time.Now()
		submissions, err = GetSubmissions(c.client, URL, n)
		if err != nil {
			return
		}
		sio_submissions.Display(submissions, first, &maxWidth, line)
		first = false
		endCount := 0
		for _, submission := range submissions {
			if submission.End {
				endCount++
			}
		}
		if endCount == len(submissions) {
			return
		}
		sub := time.Since(st)
		if sub < time.Second {
			time.Sleep(time.Second - sub)
		}
	}
}
