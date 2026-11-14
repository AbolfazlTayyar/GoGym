# 0036 — The plan PDF is an image of the page, made in the browser

- Status: accepted
- Date: 2026-11-14
- Deciders: Abolfazl Tayyar

## Context and Problem Statement

The athlete opens a shared plan (`/p/<token>`, [0035](0035-plan-share-links.md)) and can download
it as a PDF to keep, replacing the Canva PDF coaches send today. The PDF must show Persian with
joined letters, right to left, in the app's font, and the download must work on iPhone and Android.

## Decision Drivers

- Persian text that reads correctly, on every phone.
- One tap to a real `.pdf` file, not a print dialog.
- No new server-side service.

## Considered Options

1. Render the page to an image in the browser (`html-to-image`) and put it on one PDF page (`jsPDF`)
2. A print stylesheet and the browser's print dialog, Save as PDF
3. A PDF laid out by a library from the plan's data, in Go or JS
4. Headless Chrome on the server, printing the page

## Decision Outcome

Chosen option: **1**. The browser draws the page, so the Persian text is shaped exactly as on
screen. The PDF is one page as wide as the phone layout and as tall as the plan, so it reads like
the page with no breaks; a plan taller than PDF's 14400pt page limit is scaled down whole rather
than split. Both libraries are loaded only when the button is tapped. Option 2 works on Android,
but on iPhone Save as PDF hides behind Share → Print → zooming in on the preview. Option 3 fails
on the text: most PDF libraries in Go and JS don't join Persian letters or lay out RTL. Option 4
adds a Chrome service to run and secure for one button.

### Consequences

- Good: the PDF looks exactly like the page, and needs nothing on the server.
- Bad: the PDF's text can't be selected or searched, and its file is larger than a text PDF.
- Bad: rendering depends on the phone's canvas limits; a much longer plan could exceed them.
