import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { System } from "../../types";
import { SystemGrid } from "./SystemGrid";

const mockSystems: System[] = [
  {
    id: "snes",
    name: "Super Nintendo",
    description: "16-bit home console by Nintendo (1990)",
    emulators: [{ id: "retroarch:bsnes", name: "RetroArch (bsnes)" }],
  },
  {
    id: "gba",
    name: "Game Boy Advance",
    description: "32-bit handheld by Nintendo (2001)",
    emulators: [{ id: "retroarch:mgba", name: "RetroArch (mGBA)" }],
  },
  {
    id: "psx",
    name: "PlayStation",
    description: "32-bit home console by Sony (1994)",
    emulators: [{ id: "duckstation", name: "DuckStation" }],
  },
];

describe("SystemGrid", () => {
  it("renders systems grouped by manufacturer", () => {
    render(
      <SystemGrid
        systems={mockSystems}
        selections={new Map()}
        provisions={{}}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText("Nintendo")).toBeInTheDocument();
    expect(screen.getByText("Sony")).toBeInTheDocument();
    expect(screen.getByText("Super Nintendo")).toBeInTheDocument();
    expect(screen.getByText("Game Boy Advance")).toBeInTheDocument();
    expect(screen.getByText("PlayStation")).toBeInTheDocument();
  });

  it("does not render empty manufacturer groups", () => {
    const nintendoOnly: System[] = [
      {
        id: "snes",
        name: "Super Nintendo",
        description: "16-bit home console by Nintendo (1990)",
        emulators: [{ id: "retroarch:bsnes", name: "RetroArch (bsnes)" }],
      },
    ];

    render(
      <SystemGrid
        systems={nintendoOnly}
        selections={new Map()}
        provisions={{}}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText("Nintendo")).toBeInTheDocument();
    expect(screen.queryByText("Sony")).not.toBeInTheDocument();
  });

  it("passes provisions to system cards", () => {
    render(
      <SystemGrid
        systems={mockSystems}
        selections={new Map()}
        provisions={{
          psx: [
            {
              filename: "scph5501.bin",
              description: "PlayStation BIOS (USA)",
              required: true,
              status: "missing",
            },
          ],
        }}
        onToggle={vi.fn()}
      />
    );

    expect(screen.getByText(/scph5501\.bin/)).toBeInTheDocument();
  });
});
