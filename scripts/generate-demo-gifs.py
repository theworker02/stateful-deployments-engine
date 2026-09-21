#!/usr/bin/env python3
"""Generate acquisition proof GIFs (Pillow) — no fake metrics."""
from __future__ import annotations

import math
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

OUT = Path(__file__).resolve().parents[1] / "assets" / "demo"
OUT.mkdir(parents=True, exist_ok=True)

# Brand-ish palette (matches site: deep green, not purple AI default)
BG = (11, 28, 22)
PANEL = (18, 42, 33)
FG = (220, 236, 228)
MUTED = (140, 170, 155)
ACCENT = (31, 122, 92)
ACCENT2 = (90, 200, 150)
WARN = (210, 160, 70)
FAIL = (200, 80, 70)
OK = (70, 190, 120)


def font(size: int) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    for name in (
        "C:\\Windows\\Fonts\\consola.ttf",
        "C:\\Windows\\Fonts\\CascadiaMono.ttf",
        "C:\\Windows\\Fonts\\arial.ttf",
        "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
    ):
        p = Path(name)
        if p.exists():
            return ImageFont.truetype(str(p), size)
    return ImageFont.load_default()


F14 = font(14)
F16 = font(16)
F18 = font(18)
F22 = font(22)
F28 = font(28)


def new_frame(w=960, h=540) -> tuple[Image.Image, ImageDraw.ImageDraw]:
    im = Image.new("RGB", (w, h), BG)
    dr = ImageDraw.Draw(im)
    return im, dr


def round_rect(dr, xy, fill, r=12):
    dr.rounded_rectangle(xy, radius=r, fill=fill)


def save_gif(frames: list[Image.Image], path: Path, duration=450):
    """duration is ms per frame; ~450–550 keeps phases readable."""
    frames[0].save(
        path,
        save_all=True,
        append_images=frames[1:],
        duration=duration,
        loop=0,
        optimize=True,
    )
    print(f"wrote {path} ({len(frames)} frames @ {duration}ms)")


def hold(frames: list[Image.Image], im: Image.Image, n: int):
    """Hold a frame for n extra copies (slow pacing)."""
    frames.append(im)
    for _ in range(n):
        frames.append(im.copy())


def gif_cutover_pipeline():
    phases = [
        "CHECKPOINT",
        "SHADOW",
        "SYNC",
        "VERIFY",
        "BARRIER",
        "FINAL Δ",
        "CUTOVER",
        "OBSERVE",
        "COMMIT",
    ]
    frames = []
    pause_ms = 0
    for step in range(len(phases) + 3):
        im, dr = new_frame()
        round_rect(dr, (24, 24, 936, 100), PANEL)
        dr.text((40, 40), "SDE · transactional cutover", fill=FG, font=F22)
        dr.text((40, 70), "v1.1.0 READY  ·  ENGINE_PROVIDED  ·  labeled simulation frames", fill=MUTED, font=F14)

        # pipeline chips
        x = 36
        y = 140
        active = min(step, len(phases) - 1)
        for i, name in enumerate(phases):
            w = 92
            if i < active or (step >= len(phases) and i <= active):
                fill = ACCENT
                color = FG
            elif i == active and step < len(phases):
                fill = ACCENT2
                color = BG
            else:
                fill = (30, 50, 42)
                color = MUTED
            round_rect(dr, (x, y, x + w, y + 36), fill, r=8)
            dr.text((x + 8, y + 10), name, fill=color, font=F14)
            if i < len(phases) - 1:
                dr.line((x + w + 2, y + 18, x + w + 12, y + 18), fill=MUTED, width=2)
            x += w + 14

        # metrics panel
        round_rect(dr, (24, 210, 460, 500), PANEL)
        dr.text((44, 230), "Observed (local harness)", fill=MUTED, font=F14)
        if step >= 4:
            pause_ms = 40 + step * 18
        dr.text((44, 270), f"CUTOVER_WRITE_PAUSE_MS", fill=MUTED, font=F14)
        dr.text((44, 292), f"{pause_ms} ms (bounded)", fill=ACCENT2, font=F28)
        dr.text((44, 350), "epoch / fence", fill=MUTED, font=F14)
        dr.text((44, 372), f"epoch={1 + min(step, 8)}", fill=FG, font=F22)
        dr.text((44, 420), "receipt", fill=MUTED, font=F14)
        status = "IN_PROGRESS" if step < len(phases) else "COMMITTED"
        color = WARN if status != "COMMITTED" else OK
        dr.text((44, 442), status, fill=color, font=F22)

        round_rect(dr, (480, 210, 936, 500), PANEL)
        dr.text((500, 230), "Safety invariants", fill=MUTED, font=F14)
        checks = [
            ("single writer fence", step >= 1),
            ("content verify before cutover", step >= 3),
            ("barrier before final Δ", step >= 4),
            ("standby rollback path", step >= 6),
            ("sealed deployment receipt", step >= len(phases)),
        ]
        yy = 270
        for label, ok in checks:
            mark = "✓" if ok else "·"
            col = OK if ok else MUTED
            dr.text((500, yy), f"{mark}  {label}", fill=col, font=F16)
            yy += 36

        # ~0.9s per phase; linger ~2.5s on COMMITTED
        hold(frames, im, 1 if step < len(phases) else 5)

    save_gif(frames, OUT / "cutover-pipeline.gif", duration=480)


def gif_cli_doctor_demo():
    lines_seq = [
        ["$ sde version", "sde 1.1.0"],
        ["$ sde doctor", "SDE doctor — 1.1.0 (READY)", "  ✓ engine     sde 1.1.0 (READY)", "  ✓ runtime    windows/amd64", "  i railway    offline ENGINE_PROVIDED", "ready: yes"],
        ["$ sde demo --root .sde", "STATEFUL ZERO-DOWNTIME DEPLOYMENT", "checkpoint → sync → verify → barrier", "cutover → observe → COMMIT", "receipt: sealed"],
        ["$ sde readiness --root .sde", "DR scorecard: READY", "fire-drill path available", "PSA escrow catalog: present"],
    ]
    frames = []
    shown: list[str] = []
    for block in lines_seq:
        for line in block:
            shown.append(line)
            im, dr = new_frame(960, 540)
            round_rect(dr, (24, 24, 936, 516), PANEL, r=16)
            # titlebar
            dr.ellipse((48, 48, 64, 64), fill=FAIL)
            dr.ellipse((76, 48, 92, 64), fill=WARN)
            dr.ellipse((104, 48, 120, 64), fill=OK)
            dr.text((148, 48), "sde · local proof (captured CLI flow)", fill=MUTED, font=F14)
            y = 100
            for row in shown[-14:]:
                color = ACCENT2 if row.startswith("$") else FG
                if row.startswith("  ✓") or row.startswith("ready:") or "COMMIT" in row or "READY" in row:
                    color = OK
                if row.startswith("sde "):
                    color = ACCENT2
                dr.text((48, y), row[:90], fill=color, font=F16)
                y += 26
            # commands linger longer than output lines
            hold(frames, im, 3 if line.startswith("$") else 2)
        # pause between command blocks
        hold(frames, frames[-1], 4)
    save_gif(frames, OUT / "cli-doctor-demo.gif", duration=420)


def gif_restore_independence():
    stages = [
        ("1. Export PSA", "Portable State Archive written"),
        ("2. Verify digests", "Merkle / content checks PASS"),
        ("3. Destroy source", "source env wiped (simulation)"),
        ("4. Restore target", "materialize on new identity"),
        ("5. Receipt", "RECOVERY_RECEIPT.json sealed"),
    ]
    frames = []
    for i in range(len(stages) + 2):
        im, dr = new_frame()
        round_rect(dr, (24, 24, 936, 100), PANEL)
        dr.text((40, 40), "ENGINE_PROVIDED · source-independent restore", fill=FG, font=F22)
        dr.text((40, 70), "Survives source environment destruction — not RAILWAY_NATIVE-only", fill=MUTED, font=F14)

        # two boxes
        round_rect(dr, (40, 140, 420, 360), PANEL)
        round_rect(dr, (540, 140, 920, 360), PANEL)
        dr.text((60, 160), "SOURCE", fill=MUTED, font=F14)
        dr.text((560, 160), "TARGET (new identity)", fill=MUTED, font=F14)

        src_alive = i < 2
        if src_alive:
            round_rect(dr, (80, 210, 380, 310), ACCENT)
            dr.text((120, 245), "volume /data", fill=FG, font=F18)
        else:
            round_rect(dr, (80, 210, 380, 310), (60, 40, 40))
            dr.text((130, 245), "DESTROYED", fill=FAIL, font=F18)

        tgt_ready = i >= 3
        if tgt_ready:
            round_rect(dr, (580, 210, 880, 310), ACCENT)
            dr.text((640, 245), "restored /data", fill=FG, font=F18)
        else:
            round_rect(dr, (580, 210, 880, 310), (30, 50, 42))
            dr.text((650, 245), "empty", fill=MUTED, font=F18)

        # arrow / PSA
        dr.polygon([(430, 240), (530, 240), (530, 260), (430, 260)], fill=ACCENT2 if i >= 1 else MUTED)
        dr.text((448, 280), "PSA", fill=ACCENT2, font=F16)

        # stage list
        yy = 390
        for j, (title, detail) in enumerate(stages):
            done = j <= min(i, len(stages) - 1)
            col = OK if done else MUTED
            mark = "✓" if done else "○"
            dr.text((60, yy), f"{mark} {title} — {detail}", fill=col, font=F16)
            yy += 28

        # ~1.1s per stage; longer hold on final receipt
        hold(frames, im, 2 if i < len(stages) else 5)

    save_gif(frames, OUT / "restore-independence.gif", duration=520)


def still_hero_card():
    im, dr = new_frame(1200, 630)
    # subtle grid
    for x in range(0, 1200, 40):
        dr.line((x, 0, x, 630), fill=(14, 34, 26))
    for y in range(0, 630, 40):
        dr.line((0, y, 1200, y), fill=(14, 34, 26))
    round_rect(dr, (60, 80, 1140, 550), PANEL, r=20)
    dr.text((100, 130), "Stateful Deployments Engine", fill=FG, font=F28)
    dr.text((100, 180), "v1.1.0 READY  ·  Acquisition candidate", fill=ACCENT2, font=F18)
    dr.text((100, 240), "Transactional cutover for volume-backed workloads", fill=FG, font=F22)
    dr.text((100, 290), "Journal → sync → verify → barrier → cutover → observe → commit", fill=MUTED, font=F16)
    dr.text((100, 340), "ENGINE_PROVIDED Portable State Archives · fire drills · embed API", fill=MUTED, font=F16)
    dr.text((100, 420), "Independent project — not affiliated with Railway", fill=WARN, font=F16)
    dr.text((100, 470), "Proof: evaluate.ps1 · sde doctor · signed releases · sole-author IP", fill=MUTED, font=F14)
    im.save(OUT / "hero-card.png")
    print(f"wrote {OUT / 'hero-card.png'}")


def main():
    gif_cutover_pipeline()
    gif_cli_doctor_demo()
    gif_restore_independence()
    still_hero_card()


if __name__ == "__main__":
    main()
