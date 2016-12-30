# BlogMe

A simple static site genrator that just works, with sane defaults and AMP optimized pages.

Yes, yet another static site generator.

## Usage

### Installation

Just download [latest release](https://github.com/diogok/blogme/releases) at somewhere you can execute it.

Since everything have a default, just write your content in "content" folder will give results.

### Configuration file

The software will look for a config.yml at the execution folder. If none is found, it will load default configuration, that is as follow:

```yml
source: content
output: output
postDir: post
template: template
static: static
title: Title
subtitle: Subtitle
basePath: http://localhost:8585
properties:
```

- source: Where to read the content from
- output: Where to write generated files
- postDir: Where to write individual post generated files
- template: Where to read the template from
- static: Where to copy static files from, inside the template path
- title: Website title
- subtitle: Website description
- basePath: base URL
- properties: any extra information to pass on to custom templates

### Command line arguments

When executed without parameters it will generate the site once with default configiguration or config.yml.

Full paramaters are:

./blogme --config config.yml -generate -watch -serve

Where:

- config change the config.yml to read from
- generate run the site generator once
- watch watch for changes on template and source directory (from config.yml) and run generate on change
- serve starts a server at localhost:8585 serving the output (from config.yml) folder

### Writing templates

You can write the following files for the templating, if not found it will use [default template](https://github.com/diogok/blogme/tree/master/defaultTemplate). They use golang native template.

- post.html
- post\_amp.html
- index.html
- rss.xml
- sitemap.xml

### Writing content

Write content files at the configured content directory (default to "content"), each content consists of a content.md file with markdown content and a content.yml file with metadata, such as:


_hello.md_

```markdown
Hello, world!

A am a [markdown](https://daringfireball.net/projects/markdown/) content file.
```

_hello.yml_

```yml
title: hellow world
description: our first blog post
date: 2016-12-21
```

## License 

MIT

