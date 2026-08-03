---
title: Validating Markdown with mdschema
date: 2026-08-02
status: published
priority: 2
tags:
  - go
  - markdown
author:
  name: example-author
  email: author@example.com
  homepage: https://example.com
---

# Validating Markdown with mdschema

## Introduction

This post shows how mdschema keeps blog posts consistent. The frontmatter
above is validated against the schema: `status` must be one of `draft`,
`published`, or `archived`, and every entry in `tags` must come from the
allowed list.

## Content

Change `status` to a value like `pending` and run the check again to see an
enum violation:

```bash
mdschema check examples/blog-post.md --schema examples/blog-post.mdschema.yml
```

## Conclusion

Frontmatter validation catches typos in metadata before they reach your site.
