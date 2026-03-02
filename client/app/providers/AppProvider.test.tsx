import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import React from "react";

vi.mock("api/api", () => ({
  getCfg: vi.fn(),
}));

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((res) => {
    resolve = res;
  });
  return { promise, resolve };
}

describe("AppProvider loading indicator", () => {
  beforeEach(() => {vi.resetAllMocks();}
  );

  afterEach(() => {vi.unstubAllGlobals();});

  it("Loading indicator while fetching", async () => {
    const userDef = deferred<any>();
    const cfgDef = deferred<any>();

    vi.stubGlobal("fetch", vi.fn((input: any) => {
      const url = String(input);
      if (url.includes("/apis/web/v1/user/me")) {
        return userDef.promise.then((data) => ({
          ok: true,
          json: async () => data,
        })) as any;
      }
      throw new Error(`Unexpected fetch: ${url}`);
    }));

    const { getCfg } = await import("api/api");
    (getCfg as any).mockReturnValue(cfgDef.promise);

    const { AppProvider } = await import("./AppProvider");

    render(
      <AppProvider>
        <div>APP CONTENT</div>
      </AppProvider>
    );

    expect(screen.getByRole("status", { name: /loading/i })).toBeInTheDocument();
    expect(screen.queryByText("APP CONTENT")).not.toBeInTheDocument();

    userDef.resolve({ username: "test-user" });
    cfgDef.resolve({ default_theme: "testing" });

    expect(await screen.findByText("APP CONTENT")).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.queryByRole("status", { name: /loading/i })).not.toBeInTheDocument();
    });
  });
});