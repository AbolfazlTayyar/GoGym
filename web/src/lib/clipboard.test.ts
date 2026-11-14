// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from "vitest";
import { copyPendingText } from "./clipboard";

const url = "http://192.168.1.5:5173/p/q3J9xK2mT0aLwV7nB4cD8e";

// jsdom has no execCommand; the stub records what was selected when copy ran.
function stubExecCommand(result: boolean) {
  const copied: string[] = [];
  document.execCommand = vi.fn(() => {
    const area = document.querySelector("textarea");
    copied.push(area?.value ?? "");
    return result;
  });
  return copied;
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("copyPendingText", () => {
  it("falls back to a textarea over plain http, once the text arrives, and cleans up", async () => {
    vi.stubGlobal("navigator", {});
    const copied = stubExecCommand(true);

    await copyPendingText(Promise.resolve(url));

    expect(copied).toEqual([url]);
    expect(document.querySelector("textarea")).toBeNull();
  });

  it("fails when the fallback copy is refused", async () => {
    vi.stubGlobal("navigator", {});
    stubExecCommand(false);

    await expect(copyPendingText(Promise.resolve(url))).rejects.toThrow();
    expect(document.querySelector("textarea")).toBeNull();
  });

  it("hands the clipboard a promise inside the tap where the browser can take one", async () => {
    const write = vi.fn((items: { blob: Promise<Blob> }[]) => items[0].blob.then(() => undefined));
    vi.stubGlobal("navigator", { clipboard: { write } });
    vi.stubGlobal(
      "ClipboardItem",
      class {
        blob: Promise<Blob>;
        constructor(data: Record<string, Promise<Blob>>) {
          this.blob = data["text/plain"];
        }
      },
    );
    const copied = stubExecCommand(true);

    let resolve: (text: string) => void = () => {};
    const pending = copyPendingText(
      new Promise<string>((r) => {
        resolve = r;
      }),
    );
    // The write starts before the text exists: that is what keeps it inside the tap on iOS.
    expect(write).toHaveBeenCalledOnce();
    resolve(url);
    await pending;

    const blob = await write.mock.calls[0][0][0].blob;
    expect(await blob.text()).toBe(url);
    expect(copied).toEqual([]);
  });
});
