# ply

Recursive markdown to HTML converter

## Simple

* No configuration files
* The file hierarchy you see is what you get
* Nested templates

Run `ply` in your directory of choice, and all `.md` files will be converted to `.html`. The result is stored in a folder called `ply.build`.

## Templates

Create a file with name `ply.template` and it will be applied recursively to all `.md` files.

### Example

ply.template:

```
<html>
<head><title>{{ .Title }}</title></head>
<body>

{{ .Content }}

<ul>
{{ range .Sitemap -}}
    <li><a href="{{ .Path }}">{{ .Title }}</a></li>
{{ end -}}
</ul>

</body>
</html>
```

blog/ply.template:

```
<article>
{{ .Content }}
</article>
```

All pages in the root level of the website will use the `ply.template`, and all pages under the directory `blog` will first use `blog/ply.template` and then `ply.template`.

A page stored in `blog/post.md` will be converted to `blog/post.html` and get this content:

```
<html>
<head><title>A post</title></head>
<body>

<article>
<h1>A post</h1>
<p>My post content goes here</p>
</article>

<ul>
    <li><a href="blog/post.html">A post</a></li>
</ul>

</body>
</html>
```

Page title is taken from the first heading on the page. More advanced template features are available, but will be documented later.
