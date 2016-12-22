package main

import (
	"flag"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

type Post struct {
	Slug    string
	Title   string
	Content string
}

type Posts []Post

type Site struct {
	Content Posts
}

func WritePost(config *Config, file_name string) Post {
	postContent, postErr := ioutil.ReadFile(fmt.Sprintf("%s/post.html", config.Template))
	if postErr != nil {
		panic(postErr)
	}

	postTemplate, templateErr := template.New("post").Parse(string(postContent))
	if templateErr != nil {
		panic(templateErr)
	}

	name := (file_name[0 : len(file_name)-3])
	file_content, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", config.Source, file_name))
	html_content := blackfriday.MarkdownBasic(file_content)

	post := Post{Content: string(html_content), Title: name, Slug: name}

	file, _ := os.Create(fmt.Sprintf("%s/%s/%s.html", config.Output, config.PostDir, name))
	postExecuteErr := postTemplate.Execute(file, post)
	file.Close()
	if postExecuteErr != nil {
		panic(postExecuteErr)
	}

	return post
}

func WriteIndex(config *Config, site Site) {
	indexContent, indexErr := ioutil.ReadFile(fmt.Sprintf("%s/index.html", config.Template))
	if indexErr != nil {
		panic(indexErr)
	}
	indexTemplate, indexTemplateErr := template.New("index").Parse(string(indexContent))
	if indexTemplateErr != nil {
		panic(indexTemplateErr)
	}

	indexFile, _ := os.Create(fmt.Sprintf("%s/index.html", config.Output))
	indexExecuteErr := indexTemplate.Execute(indexFile, site)
	indexFile.Close()
	if indexExecuteErr != nil {
		panic(indexExecuteErr)
	}
}

func WriteRss(config *Config, site Site) {
	rssContent, rssErr := ioutil.ReadFile(fmt.Sprintf("%s/rss.xml", config.Template))
	if rssErr != nil {
		panic(rssErr)
	}
	rssTemplate, rssTemplateErr := template.New("rss").Parse(string(rssContent))
	if rssTemplateErr != nil {
		panic(rssTemplateErr)
	}

	rssFile, _ := os.Create(fmt.Sprintf("%s/rss.xml", config.Output))
	rssExecuteErr := rssTemplate.Execute(rssFile, site)
	rssFile.Close()
	if rssExecuteErr != nil {
		panic(rssExecuteErr)
	}
}

func Generate(config *Config) {
	files, srcErr := ioutil.ReadDir(config.Source)
	if srcErr != nil {
		panic(srcErr)
	}

	items := Posts{}
	for _, file := range files {
		file_name := file.Name()
		if string(file_name[0]) == "." {
			continue
		}

		post := WritePost(config, file_name)

		items = append(items, post)
	}

	WriteIndex(config, Site{items})
	WriteRss(config, Site{items})
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "config.yml", "Location of config file")
	flag.Parse()

	config := LoadConfig(configFile)
	log.Println(config)
	Generate(config)
}
