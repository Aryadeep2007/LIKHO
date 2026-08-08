# Likho (लिखो) ✍️

> **A blog that comes in a single file.**

No database. No cloud. No signup. No hosting. No internet required.

Just run Likho, write in Markdown, and your blog is ready.

And if you're on the same Wi-Fi, scan the QR code and open the blog from your phone — **no domain, no hosting, no internet.**

---

## ✨ What is Likho?

Likho is a **self-contained, local-first blogging platform** built in Go.

Your entire blog lives as ordinary Markdown files on your machine. The application itself contains the web server, editor, templates, CSS, and JavaScript in a small standalone binary.

```text
likho.exe
    │
    ├── Web server
    ├── Markdown editor
    ├── Blog renderer
    ├── Search
    ├── Graph
    ├── Analytics
    └── HTML / CSS / JS
```

There is no database sitting behind it.

Your posts are just files.

That means you can open them in Notepad, back them up with Git, sync them with Dropbox, or move them to another machine without worrying about database exports.

---

## 🚀 Quick Start

### Windows

1. Download and extract the Windows package.
2. Double-click `run.bat`.
3. Open the displayed local address.

That's it.

Nothing needs to be installed.

The Windows build contains a single `likho.exe` with everything required to run the application.

### Linux / macOS

```bash
./run.sh
```

### Docker

```bash
docker compose up
```

Then open:

```text
http://localhost:4000
```

---

## 📱 Share Your Blog Over Wi-Fi

Likho isn't restricted to `localhost`.

It can listen on your local network, allowing other devices connected to the same Wi-Fi network to access the blog.

Open the **QR code** from the header, scan it with a phone, and the blog opens immediately.

```text
Laptop
  │
  │  Same Wi-Fi
  │
  ├────────────── Phone
  │
  ├────────────── Tablet
  │
  └────────────── Another Laptop
```

No:

* Hosting
* Domain
* Internet connection
* Cloud account
* External server

Perfect for classrooms, hackathons, home networks, workshops, and local communities.

> ⚠️ **Important:** Anyone who can access the shared address can edit the blog. Don't expose Likho on an untrusted network.

---

## 🧠 Features

### 🔗 Wikilinks

Connect your ideas using `[[wikilinks]]`.

Write:

```markdown
I learned this while working on [[My First Project]].
```

Likho automatically turns it into a link.

If the linked post doesn't exist yet, clicking it takes you directly to the editor so you can create it.

---

### 🕸️ Knowledge Graph

The **Graph / Naksha** view visualizes the connections between your posts.

Instead of treating your blog as a collection of isolated articles, Likho lets you build a network of ideas.

```text
       ┌──────────────┐
       │  Go Basics   │
       └──────┬───────┘
              │
       ┌──────▼───────┐
       │ Web Servers  │
       └──────┬───────┘
              │
       ┌──────▼───────┐
       │    Likho     │
       └──────────────┘
```

Inspired by the idea of connected notes popularized by tools such as Obsidian.

---

### 👀 Reading Depth

Traditional analytics might tell you:

> "72% of visitors read this post."

Likho tries to answer a more useful question:

> **"Where did they stop reading?"**

Every post has a small reading-depth visualization showing how far readers got through the content.

The data is stored alongside the post itself rather than being sent to a third-party analytics service.

No:

* Cookies
* Tracking IDs
* Google Analytics
* External analytics service

Just local data.

---

## 🛠️ More Features

Likho also includes:

* ⌨️ `Ctrl+K` command palette
* 🔍 Full-text search
* 🌙 Dark mode
* 📝 Markdown editing
* 🔄 Revision history
* ↩️ Restore previous revisions
* 🧠 TF-IDF related posts
* 🏷️ Automatic tag suggestions
* ⏱️ Reading-time calculation
* 🔥 Writing streaks
* 📦 `.likho` backup/export format
* 📱 Mobile-friendly interface
* 🔗 Wikilinks
* 🕸️ Knowledge graph
* 📡 LAN sharing
* 📷 QR-code access

---

## 📂 Your Data Stays With You

Posts are stored in:

```text
blog-data/
└── posts/
    ├── first-post.md
    ├── my-project.md
    └── notes.md
```

A post is just a Markdown file with metadata:

```markdown
---
title: My First Post
tags: go, hackathon
created: 2026-08-07T10:30:00Z
published: true
views: 12
depth: 8,8,7,5,5,4,3,3,2,2
---

Markdown starts here...
```

You can:

* Open it in Notepad or any text editor
* Commit it to Git
* Back it up to Dropbox
* Copy it to another computer
* Edit it outside Likho
* Keep your data even if the application is removed

**Your writing isn't trapped inside a database.**

---

## 🏗️ Architecture

Likho is intentionally small.

The repository is organized roughly like this:

```text
LIKHO/
├── main.go              # Server startup, routes and templates
├── store.go             # Post storage and revisions
├── handlers.go          # HTTP handlers
├── render.go            # Markdown + wikilink rendering
├── search.go            # Search, tags and related posts
├── netinfo.go           # LAN information and QR generation
├── portable.go          # .likho export / import
├── seed.go              # Initial sample content
│
├── web/                 # HTML templates, CSS and JavaScript
│
├── dist/                # Built binaries/packages
├── docs/                # Documentation
│
├── build.sh             # Cross-platform build script
├── run.sh               # Linux/macOS launcher
├── run.bat              # Windows launcher
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

The web assets are embedded into the application binary, allowing Likho to be distributed as a self-contained application.

---

## ⚙️ Technology

Likho is built primarily with **Go** and its standard library.

Key dependencies include:

* **[Goldmark](https://github.com/yuin/goldmark)** — Markdown parsing and rendering
* **[go-qrcode](https://github.com/skip2/go-qrcode)** — QR code generation

The goal is to keep the dependency footprint small and the resulting binary portable.

The project can be built with:

```text
CGO_ENABLED=0
```

---

## 🔨 Build From Source

You can build Likho using the provided build script.

### Build everything

```bash
./build.sh
```

This builds the supported platforms and creates packages in:

```text
dist/
```

### Build Windows only

```bash
./build.sh windows
```

The resulting Windows executable can be distributed without requiring a Go installation on the target machine.

### Docker

If you prefer not to install Go locally:

```bash
docker compose up
```

---

## 🐳 Docker

Build and run Likho with Docker:

```bash
docker compose up --build
```

Then visit:

```text
http://localhost:4000
```

Your blog data should be kept in the mounted data directory so that containers can be recreated without losing your posts.

---

## 🔐 Privacy & Security

Likho is designed to be **local-first**.

It doesn't require:

* User accounts
* Cloud storage
* Third-party analytics
* Tracking cookies
* An internet connection

However, local-first does **not** automatically mean secure on an untrusted network.

### No Authentication

There is currently no login system.

Anyone who can reach your Likho instance may be able to edit it.

**Use it on trusted networks.**

A home, classroom, or private LAN is appropriate.

An open café or public Wi-Fi network is not.

### Raw HTML

Raw executable HTML such as:

```html
<script>
```

is intentionally blocked.

This is particularly important because Likho can expose the editor over a local network.

### Images

There is currently no built-in image upload system.

External image URLs can be used from Markdown.

---

## ⚠️ Current Limitations

Likho is intentionally simple, but there are a few things to know:

* No authentication or user accounts
* No built-in image uploads
* No multi-user editing
* No conflict resolution for simultaneous edits
* Raw executable HTML is blocked
* LAN access should only be enabled on trusted networks

If two people edit the same post at the same time, the version saved later can overwrite the earlier one.

---

## 🎯 Why Likho?

Modern blogging platforms often come with a lot of infrastructure:

```text
Database
    ↓
Backend
    ↓
Authentication
    ↓
Cloud hosting
    ↓
Domain
    ↓
Analytics
    ↓
Deployment
    ↓
Blog
```

Likho asks:

> **What if we removed almost all of that?**

```text
Markdown files
      +
  One binary
      ↓
    Blog
```

That's the idea.

**Your blog should be yours.**

---

## 🗺️ Roadmap

Some things that could be explored in future versions:

* [ ] Authentication / access control
* [ ] Better multi-user support
* [ ] Conflict detection
* [ ] Image management
* [ ] Improved export/import
* [ ] More themes
* [ ] Better graph visualization
* [ ] More powerful Markdown extensions
* [ ] Automated backups

---

## 🤝 Contributing

Contributions, ideas, bug reports, and improvements are welcome.

To get started:

```bash
git clone https://github.com/Aryadeep2007/LIKHO.git
cd LIKHO
```

Make your changes, test them locally, and open a pull request.

If you find a bug or have an idea for a feature, feel free to open an issue.

---

## 📄 License

See the repository for the current license information.

---

## 💡 Philosophy

Likho isn't trying to replace WordPress, Medium, Ghost, or every other blogging platform.

It's trying to answer a smaller question:

> **Can a useful blog just be a file?**

Turns out, pretty much.

**Write. Save. Share. That's it.** ✍️
