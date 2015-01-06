package schema

import (
	"bytes"
	"mime/multipart"
	"strconv"

	. "gopkg.in/check.v1"
)

type multipartDecoder struct{}

type fileInfo struct {
	fieldName string
	fileName  string
	data      string
}

var _ = Suite(&multipartDecoder{})

func (s *multipartDecoder) Test_SingleFile(c *C) {
	expectedBlogpost := BlogPost{
		Post: Post{Title: "Glorious Post Title"}, Id: 1,
		Author:  Person{Name: "Michael Boke"},
		Ratings: []int{3, 5, 4},
	}
	blogPost := BlogPost{}
	mf := buildMultipartWithFile(
		[]fileInfo{
			fileInfo{
				fieldName: "headerImage",
				fileName:  "message.txt",
				data:      "Whoohoo yippie yeey lol :)",
			},
		}, &expectedBlogpost)
	decoder := NewMultipartDecoder()
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(&blogPost, mf)

	c.Assert(err, IsNil)

	c.Assert(blogPost.HeaderImage, NotNil)
	c.Assert(blogPost.HeaderImage.Filename, Equals, "message.txt")
	c.Assert(unpackFileData(blogPost.HeaderImage), Equals, "Whoohoo yippie yeey lol :)")
	c.Assert(blogPost.Pictures, HasLen, 0)

	blogPost.HeaderImage = nil //< no need to check this one twice
	c.Assert(blogPost, DeepEquals, expectedBlogpost)
}

func (s *multipartDecoder) Test_MultipleFiles(c *C) {
	expectedBlogpost := BlogPost{
		Post: Post{Title: "Yes now with a pointer"}, Id: 40,
		Author:       Person{Name: "Michael Boke"},
		Coauthor:     &Person{Name: "The other guy"},
		Readers:      []Person{Person{Name: "Person a"}, Person{Name: "Person b"}},
		Contributors: []*Person{&Person{Name: "Michael Boke"}, &Person{Name: "The other guy"}},
	}
	blogPost := BlogPost{}
	mf := buildMultipartWithFile([]fileInfo{
		fileInfo{
			fieldName: "picture",
			fileName:  "cool-gopher-fact.txt",
			data:      "Did you know? https://plus.google.com/+MatthewHolt/posts/GmVfd6TPJ51",
		},
		fileInfo{
			fieldName: "picture",
			fileName:  "gophercon2014.txt",
			data:      "@bradfitz has a Go time machine: https://twitter.com/mholt6/status/459463953395875840",
		},
	}, &expectedBlogpost)

	decoder := NewMultipartDecoder()
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(&blogPost, mf)

	c.Assert(err, IsNil)

	c.Assert(blogPost.HeaderImage, IsNil)
	c.Assert(blogPost.Pictures, HasLen, 2)
	c.Assert(blogPost.Pictures[0].Filename, Equals, "cool-gopher-fact.txt")
	c.Assert(unpackFileData(blogPost.Pictures[0]), Equals, "Did you know? https://plus.google.com/+MatthewHolt/posts/GmVfd6TPJ51")
	c.Assert(blogPost.Pictures[1].Filename, Equals, "gophercon2014.txt")
	c.Assert(unpackFileData(blogPost.Pictures[1]), Equals, "@bradfitz has a Go time machine: https://twitter.com/mholt6/status/459463953395875840")

	blogPost.Pictures = nil //< no need to check this one twice
	c.Assert(blogPost, DeepEquals, expectedBlogpost)
}

func (s *multipartDecoder) Test_SingleFileAndMultipleFiles(c *C) {
	blogPost := BlogPost{}
	mf := buildMultipartWithFile([]fileInfo{
		fileInfo{
			fieldName: "headerImage",
			fileName:  "social media.txt",
			data:      "Hey, you should follow @mholt6 (Twitter) or +MatthewHolt (Google+)",
		},
		fileInfo{
			fieldName: "picture",
			fileName:  "thank you!",
			data:      "Also, thanks to all the contributors of this package!",
		},
		fileInfo{
			fieldName: "picture",
			fileName:  "btw...",
			data:      "This tool translates JSON into Go structs: http://mholt.github.io/json-to-go/",
		},
	}, nil)

	decoder := NewMultipartDecoder()
	decoder.IgnoreUnknownKeys(true)
	err := decoder.Decode(&blogPost, mf)

	c.Assert(err, IsNil)

	c.Assert(blogPost.HeaderImage, NotNil)
	c.Assert(blogPost.HeaderImage.Filename, Equals, "social media.txt")
	c.Assert(unpackFileData(blogPost.HeaderImage), Equals, "Hey, you should follow @mholt6 (Twitter) or +MatthewHolt (Google+)")

	c.Assert(blogPost.Pictures, HasLen, 2)
	c.Assert(blogPost.Pictures[0].Filename, Equals, "thank you!")
	c.Assert(unpackFileData(blogPost.Pictures[0]), Equals, "Also, thanks to all the contributors of this package!")
	c.Assert(blogPost.Pictures[1].Filename, Equals, "btw...")
	c.Assert(unpackFileData(blogPost.Pictures[1]), Equals, "This tool translates JSON into Go structs: http://mholt.github.io/json-to-go/")
}

func buildMultipartWithFile(files []fileInfo, blogPost *BlogPost) *multipart.Form {
	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)

	//append files
	for _, file := range files {
		formFile, err := w.CreateFormFile(file.fieldName, file.fileName)
		if err != nil {
			panic("Could not create FormFile (multiple files): " + err.Error())
		}
		formFile.Write([]byte(file.data))
	}

	//append fields
	if blogPost != nil {
		w.WriteField("title", blogPost.Title)
		w.WriteField("content", blogPost.Content)
		w.WriteField("id", strconv.Itoa(blogPost.Id))
		w.WriteField("ignored", blogPost.Ignored)
		for _, value := range blogPost.Ratings {
			w.WriteField("rating", strconv.Itoa(value))
		}
		w.WriteField("author.name", blogPost.Author.Name)
		w.WriteField("author.email", blogPost.Author.Email)

		if blogPost.Coauthor != nil {
			w.WriteField("coauthor.name", blogPost.Coauthor.Name)
			w.WriteField("coauthor.email", blogPost.Coauthor.Email)
		}

		for key, value := range blogPost.Contributors {
			w.WriteField("contributors."+strconv.Itoa(key)+".name", value.Name)
			w.WriteField("contributors."+strconv.Itoa(key)+".email", value.Email)
		}

		for key, value := range blogPost.Readers {
			w.WriteField("readers."+strconv.Itoa(key)+".name", value.Name)
			w.WriteField("readers."+strconv.Itoa(key)+".email", value.Email)
		}
	}

	err := w.Close()
	if err != nil {
		panic("Could not close multipart writer: " + err.Error())
	}

	r := multipart.NewReader(b, w.Boundary())
	mf, err := r.ReadForm(1024 * 1024 * 16)
	if err != nil {
		panic("Could not read multipart: " + err.Error())
	}

	return mf
}

func unpackFileData(fh *multipart.FileHeader) string {
	if fh == nil {
		return ""
	}

	f, err := fh.Open()
	if err != nil {
		panic("Could not open file header:" + err.Error())
	}
	defer f.Close()

	var fb bytes.Buffer
	_, err = fb.ReadFrom(f)
	if err != nil {
		panic("Could not read from file header:" + err.Error())
	}

	return fb.String()
}
