package schema

import (
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

// These types are mostly contrived examples, but they're used
// across many test cases. The idea is to cover all the scenarios
// that this binding package might encounter in actual use.
type (
	// For basic test cases with a required field
	Post struct {
		Title   string `schema:"title" json:"title" binding:"Required"`
		Content string `schema:"content" json:"content"`
	}

	// To be used as a nested struct (with a required field)
	Person struct {
		Name  string `schema:"name" json:"name" binding:"Required"`
		Email string `schema:"email" json:"email"`
	}

	// For advanced test cases: multiple values, embedded
	// and nested structs, an ignored field, and single
	// and multiple file uploads
	BlogPost struct {
		Post
		Id           int                     `schema:"id" binding:"Required"`
		Ignored      string                  `schema:"-" json:"-"`
		Ratings      []int                   `schema:"rating" json:"ratings"`
		Author       Person                  `json:"author"`
		Coauthor     *Person                 `json:"coauthor"`
		Readers      []Person                `schema:"readers"`
		Contributors []*Person               `schema:"contributors"`
		HeaderImage  *multipart.FileHeader   `schema:"headerImage"`
		Pictures     []*multipart.FileHeader `schema:"picture"`
		unexported   string                  `schema:"unexported"`
	}

	EmbedPerson struct {
		*Person
	}

	SadForm struct {
		AlphaDash    string   `schema:"AlphaDash" binding:"AlphaDash"`
		AlphaDashDot string   `schema:"AlphaDashDot" binding:"AlphaDashDot"`
		MinSize      string   `schema:"MinSize" binding:"MinSize(5)"`
		MinSizeSlice []string `schema:"MinSizeSlice" binding:"MinSize(5)"`
		MaxSize      string   `schema:"MaxSize" binding:"MaxSize(1)"`
		MaxSizeSlice []string `schema:"MaxSizeSlice" binding:"MaxSize(1)"`
		Email        string   `schema:"Email" binding:"Email"`
		Url          string   `schema:"Url" binding:"Url"`
		UrlEmpty     string   `schema:"UrlEmpty" binding:"Url"`
		Range        int      `schema:"Range" binding:"Range(1,2)"`
		RangeInvalid int      `schema:"RangeInvalid" binding:"Range(1)"`
		In           string   `schema:"In" binding:"Default(0);In(1,2,3)"`
		InInvalid    string   `schema:"InInvalid" binding:"In(1,2,3)"`
		NotIn        string   `schema:"NotIn" binding:"NotIn(1,2,3)"`
		Include      string   `schema:"Include" binding:"Include(a)"`
		Exclude      string   `schema:"Exclude" binding:"Exclude(a)"`
	}
)

func newRequest(method, query, body, contentType string) *http.Request {

	var bodyReader io.Reader
	if body != "-nil-" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, query, bodyReader)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", contentType)
	return req
}

const (
	formContentType = "application/x-www-form-urlencoded"
)
