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
	Content    string
	Properties map[string]string
	Config     Config
}

type Posts []Post

type Site struct {
	Content Posts
	Config  Config
}

func WritePost(config *Config, file_name string) Post {
	name := (file_name[0 : len(file_name)-3])
	file_content, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", config.Source, file_name))
	html_content := blackfriday.MarkdownBasic(file_content)

	post := Post{Content: string(html_content), Slug: name, Config: *config}

	_, err := os.Stat(fmt.Sprintf("%s/%s.yml", config.Source, name))
	if !os.IsNotExist(err) {
		recontent, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s.yml", config.Source, name))
		yaml.Unmarshal(recontent, &post.Properties)
	}

	postContent, postErr := ioutil.ReadFile(fmt.Sprintf("%s/post.html", config.Template))
	if postErr != nil {
		log.Println(postErr)
		log.Println("Using default")
		postContent, _ = Asset("defaultTemplate/post.html")
	}

	postTemplate, templateErr := template.New("post").Parse(string(postContent))
	if templateErr != nil {
		log.Println(templateErr)
	}

	file, _ := os.Create(fmt.Sprintf("%s/%s/%s.html", config.Output, config.PostDir, name))
	postExecuteErr := postTemplate.Execute(file, post)
	file.Close()
	if postExecuteErr != nil {
		log.Println(postExecuteErr)
	}

	postAmpContent, postAmpErr := ioutil.ReadFile(fmt.Sprintf("%s/post_amp.html", config.Template))
	if postAmpErr != nil {
		log.Println("Using AMP default template")
		postAmpContent, _ = Asset("defaultTemplate/post_amp.html")
	}

	postAmpTemplate, templateAmpErr := template.New("post").Parse(string(postAmpContent))
	if templateAmpErr != nil {
		log.Println(templateAmpErr)
	}

	ampFile, _ := os.Create(fmt.Sprintf("%s/%s/%s-amp.html", config.Output, config.PostDir, name))
	postAmpExecuteErr := postAmpTemplate.Execute(ampFile, post)
	file.Close()
	if postAmpExecuteErr != nil {
		log.Println(postAmpExecuteErr)
	}

	return post
}

func WriteSite(name string, config *Config, site Site) {
	indexContent, indexErr := ioutil.ReadFile(fmt.Sprintf("%s/%s", config.Template, name))
	if indexErr != nil {
		log.Println(indexErr)
		log.Println("Using default")
		indexContent, _ = Asset(fmt.Sprintf("defaultTemplate/%s", name))
	}
	indexTemplate, indexTemplateErr := template.New("index").Parse(string(indexContent))
	if indexTemplateErr != nil {
		log.Println(indexTemplateErr)
	}

	indexFile, _ := os.Create(fmt.Sprintf("%s/%s", config.Output, name))
	indexExecuteErr := indexTemplate.Execute(indexFile, site)
	indexFile.Close()
	if indexExecuteErr != nil {
		log.Println(indexExecuteErr)
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
		log.Println(srcErr)
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

	WriteSite("index.html", config, Site{items, *config})
	WriteSite("rss.xml", config, Site{items, *config})
	WriteSite("site.xml", config, Site{items, *config})

	CopyStatic(config)
}

func Watch(config *Config, waiter *sync.WaitGroup) {
	log.Println("Watching")
	waiter.Add(1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}
	err = watcher.Add(config.Template)
	if err != nil {
		log.Println(err)
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

	if serve {
		config.BasePath = "http://localhost:8585"
	}

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
