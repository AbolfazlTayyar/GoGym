// Copies text that is still being fetched. iOS only lets a page copy during the tap, and waiting for a
// request ends that window, so the copy starts at once with a promise of the text where the browser can.
export async function copyPendingText(text: Promise<string>): Promise<void> {
  // navigator.clipboard exists only on https; plain http on the LAN goes to the fallback.
  if (navigator.clipboard?.write && typeof ClipboardItem !== "undefined") {
    const blob = text.then((t) => new Blob([t], { type: "text/plain" }));
    try {
      await navigator.clipboard.write([new ClipboardItem({ "text/plain": blob })]);
      return;
    } catch {
      // A browser that can't take a promise here still gets the fallback.
    }
  }
  copyWithTextarea(await text);
}

// The pre-clipboard-API way, which works over http on iOS and Android: select a hidden textarea and copy.
export function copyWithTextarea(text: string) {
  const area = document.createElement("textarea");
  area.value = text;
  // readOnly keeps the keyboard down; fixed keeps the page from scrolling to it.
  area.readOnly = true;
  area.style.position = "fixed";
  area.style.opacity = "0";
  document.body.append(area);
  area.select();
  // iOS ignores select() on a textarea; it needs the range.
  area.setSelectionRange(0, text.length);
  const copied = document.execCommand("copy");
  area.remove();
  if (!copied) throw new Error("copy failed");
}
