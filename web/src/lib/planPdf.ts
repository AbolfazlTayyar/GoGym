// PDF's largest page side, in points (about 5 m).
const maxPagePt = 14400;

// Characters phones and WhatsApp refuse in a file name.
const unsafeFileChars = /[\\/:*?"<>|]/g;

export function planPdfFileName(firstName: string, lastName: string, title: string): string {
  return `${firstName} ${lastName} - ${title} - gogymapp.pdf`.replace(unsafeFileChars, "-");
}

// The browser draws the page, so Persian letters join and run right to left; the PDF is one image of it
// on one page as wide as the phone layout (docs/adr/0036). Both libraries load only when asked for.
export async function downloadPlanPdf(node: HTMLElement, fileName: string): Promise<void> {
  const [{ toJpeg }, { jsPDF }] = await Promise.all([import("html-to-image"), import("jspdf")]);
  // Without the web fonts loaded, the image falls back to a system font.
  await document.fonts.ready;

  const background = getComputedStyle(document.documentElement).getPropertyValue("--pageEnd");
  const image = await toJpeg(node, {
    pixelRatio: 2,
    quality: 0.92,
    backgroundColor: background.trim() || "#fff",
  });

  // One CSS pixel is one point; a plan taller than PDF allows is scaled down whole rather than split.
  const scale = Math.min(1, maxPagePt / node.offsetHeight);
  const width = node.offsetWidth * scale;
  const height = node.offsetHeight * scale;
  const pdf = new jsPDF({ unit: "pt", format: [width, height], orientation: "portrait" });
  pdf.addImage(image, "JPEG", 0, 0, width, height);
  pdf.save(fileName);
}
