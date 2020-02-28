package models

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

type Ebook struct {
	URL        string `json:"url"`
	Title      string `json:"title"`
	Image      string `json:"image"`
	Trending   bool   `json:"trending"`
	Rate       string `json:"rate"`
	View       string `json:"view"`
	Status     string `json:"status"`
	Writer     string `json:"writer"`
	WriterLink string `json:"writer_link"`
	Categories string `json:"categories"`
}

type Ebooks struct {
	TotalPages  int     `json:"total_pages"`
	TotalEbooks int     `json:"total_ebooks"`
	List        []Ebook `json:"ebooks"`
}

func NewEbooks() *Ebooks {
	return &Ebooks{}
}

func (ebooks *Ebooks) GetTotalPages(url string) error {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return err
	}
	lastPageLink, _ := doc.Find("ul li.nexts a").Attr("href")
	if lastPageLink == "javascript:void();" {
		ebooks.TotalPages = 1
		return nil
	}
	split := strings.Split(lastPageLink, "/trang-")
	totalPages, _ := strconv.Atoi(split[1])
	ebooks.TotalPages = totalPages
	return nil
}

func (ebooks *Ebooks) getEbooksByUrl(url string, db *sql.DB) error {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return err
	}

	doc.Find("div.table-list.pc > table > tbody > tr").Each(func(i int, s *goquery.Selection) {
		docTitle, exists := s.Find("h3.rv-home-a-title a").Attr("title")
		if !exists {
			docTitle = ""
		}
		docImg, exists := s.Find("img.image-book").Attr("src")
		if !exists {
			docImg = "#"
		}
		docLink, exists := s.Find("h3.rv-home-a-title a").Attr("href")
		if !exists {
			docLink = "#"
		}
		docRate := s.Find("div.rate").Text()
		docView := s.Find("div.view").Text()
		docStatus := s.Find("td.info > p:nth-child(4)").Text()
		docWriter := s.Find("td.info > p:nth-child(5) > a").Text()
		docWriterLink, exists := s.Find("td.info > p:nth-child(5) > a").Attr("href")
		if !exists {
			docWriterLink = "#"
		}
		docCategories := s.Find("  td.info > p:nth-child(6)").Text()
		Ebook := Ebook{
			URL:        docLink,
			Title:      docTitle,
			Image:      docImg,
			Trending:   true,
			Rate:       docRate,
			View:       docView,
			Status:     docStatus,
			Writer:     docWriter,
			WriterLink: docWriterLink,
			Categories: docCategories,
		}
		ebooks.TotalEbooks++

		// query := "INSERT INTO novel (title, description, image, category_id, writer, writerlink, rate, view, status)  "
		// query := "INSERT INTO novel (title, description, image, category_id, writer, writerlink, rate, view, status) values (?, " +
		// docLink + ",'" + docImg + "'," + docCategories + "," + docWriter + "," + docWriterLink + "," + docRate + "," + docView + "," + docStatus + ")"
		insertToTable(db, &Ebook)
		ebooks.List = append(ebooks.List, Ebook)
	})
	return nil
}
func (ebooks *Ebooks) GetAllEbooks(currentUrl string) error {
	db, err := sql.Open("mysql", DB_USER+":"+DB_PASS+"@/"+DB_NAME+"?charset="+DB_CHARSET)
	if err != nil {
		log.Fatal("Cannot open DB connection", err)
	}
	eg := errgroup.Group{}
	if ebooks.TotalPages > 0 {
		for i := 1; i <= ebooks.TotalPages; i++ {
			uri := fmt.Sprintf("%vtrang-%v", currentUrl, i)
			eg.Go(func() error {
				err := ebooks.getEbooksByUrl(uri, db)
				if err != nil {
					return err
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			return err
		}
	}
	return nil
}

const (
	DB_USER = "novel_admin"
	// DB_PASS    = "123qwe123"
	DB_PASS    = "123Qwe!23"
	DB_NAME    = "novel_db"
	DB_CHARSET = "utf8"
	DB_HOST    = "api.xuanlocle.com"
)

func insertToTable(db *sql.DB, ebook *Ebook) {

	// query := "INSERT INTO `novel` (`title`, `description`, `image`, `category_id`, `writer`, `writerlink`, `rate`, `view`, `status`) VALUES (?,?,?,?,?,?,?,?,?)"

	// // stmt, err := db.Prepare("INSERT " + tableName + " SET name=?, created_at= NOW(), created_by= 'novel_admin', updated_at = NOW(), last_updated_by= 'novel_admin'")
	// stmt, err := db.Prepare(query)
	// if err != nil {
	// 	log.Fatal("Cannot prepare DB statement", err)
	// }

	// res, err := stmt.Exec(ebook.Title, ebook.URL, ebook.Image, ebook.Categories, ebook.Writer, ebook.WriterLink, ebook.Rate, ebook.View, ebook.Status)
	// if err != nil {
	// 	log.Fatal("Cannot run insert statement", err)
	// }
	// id, _ := res.LastInsertId()
	// fmt.Printf("Inserted row %d: %s", id, ebook.Title)
}
