package links

import (
	"itadakimasu-dl/config"
	"itadakimasu-dl/interfaces"
	"itadakimasu-dl/network"
	"math/rand/v2"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type stape struct {
	Url  string
	Id   string
	Name string
	Conf *config.Link
}

func (s *stape) GetConfig() *config.Link {
	return s.Conf
}

func (s *stape) Download(outputFile string) error {
	time.Sleep(time.Duration(rand.IntN(1000)))
	doc, err := network.HttpGetDocument(s.Url)
	if err != nil {
		return err
	}
	regex := regexp.MustCompile(`&expires=\d+&ip=[^&]+&token=[^'")]+`)
	doc.Find("script").Each(func(i int, sel *goquery.Selection) {
		scriptContent := sel.Text()
		matches := regex.FindAllString(scriptContent, -1)
		for _, match := range matches {
			videoLink := "https://streamtape.com/get_video?id=" + s.Id + match + "&stream=1"
			err = network.DownloadFile(videoLink, outputFile, "")
			break
		}

	})
	return err
}

func (s *stape) GetId() string {
	return s.Id
}

func (s *stape) GetServerName() string {
	return s.Name
}

func (s *stape) GetUrl() string {
	return s.Url
}

func extractStreamtapeID(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	p := strings.Trim(u.Path, "/")
	parts := strings.Split(p, "/")
	if len(parts) >= 2 && (parts[0] == "v" || parts[0] == "e") {
		return parts[1]
	}
	if len(parts) >= 1 && parts[0] != "" && parts[0] != "v" && parts[0] != "e" {
		return parts[0]
	}
	if id := u.Query().Get("id"); id != "" {
		return id
	}
	return ""
}

func NewStapeLink(rawURL string, name string, conf *config.Link) interfaces.ILink {
	id := extractStreamtapeID(rawURL)
	return &stape{
		Url:  rawURL,
		Id:   id,
		Name: name,
		Conf: conf,
	}
}
