package main

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/ghodss/yaml"
	"github.com/russross/blackfriday"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/template"
)

type Post struct {
	Slug       string
	Title      string
	Content    string
	Properties map[string]string
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

	_, err := os.Stat(fmt.Sprintf("%s/%s.yml", config.Source, name))
	if !os.IsNotExist(err) {
		recontent, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s.yml", config.Source, name))
		yaml.Unmarshal(recontent, &post.Properties)
	}

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

func CopyDir(from string, to string) {
	os.MkdirAll(to, 0755)

	files, _ := ioutil.ReadDir(from)
	for _, file := range files {
		file_name := file.Name()
		if strings.HasPrefix(file_name, ".") {
			continue
		}
		if file.IsDir() {
			CopyDir(fmt.Sprintf("%s/%s", from, file_name), fmt.Sprintf("%s/%s", to, file_name))
		} else {
			in, _ := os.Open(fmt.Sprintf("%s/%s", from, file_name))
			out, _ := os.Create(fmt.Sprintf("%s/%s", to, file_name))
			io.Copy(out, in)
			in.Close()
			out.Close()
		}
	}
}

func CopyStatic(config *Config) {
	CopyDir(fmt.Sprintf("%s/%s", config.Template, config.Static), fmt.Sprintf("%s/%s", config.Output, config.Static))
}

func Generate(config *Config) {
	log.Println("Generating")

	files, srcErr := ioutil.ReadDir(config.Source)
	if srcErr != nil {
		panic(srcErr)
	}

	items := Posts{}
	for _, file := range files {
		file_name := file.Name()
		if strings.HasPrefix(file_name, ".") || !strings.HasSuffix(file_name, ".md") {
			continue
		}

		post := WritePost(config, file_name)

		items = append(items, post)
	}

	WriteIndex(config, Site{items})
	WriteRss(config, Site{items})
	CopyStatic(config)
}

func Watch(config *Config, waiter *sync.WaitGroup) {
	log.Println("Watching")
	waiter.Add(1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					Generate(config)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
		waiter.Done()
	}()

	err = watcher.Add(config.Source)
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add(config.Template)
	if err != nil {
		log.Fatal(err)
	}
}

func Serve(config *Config, waiter *sync.WaitGroup) {
	log.Println("Serving at http://localhost:8585")
	waiter.Add(1)
	http.ListenAndServe(":8585", http.FileServer(http.Dir(config.Output)))
	waiter.Done()
}

func main() {
	var configFile string

	var generate bool
	var watch bool
	var serve bool

	flag.StringVar(&configFile, "config", "config.yml", "Location of config file")
	flag.BoolVar(&generate, "generate", true, "Command to generate the blog")
	flag.BoolVar(&watch, "watch", false, "Watch for changes and generate when needed")
	flag.BoolVar(&serve, "serve", false, "Start a server on generated site")

	flag.Parse()

	config := LoadConfig(configFile)

	var waiter sync.WaitGroup

	if generate {
		Generate(config)
	}

	if watch {
		Watch(config, &waiter)
	}

	if serve {
		Serve(config, &waiter)
	}

	waiter.Wait()
}
