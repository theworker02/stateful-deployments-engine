# Raster exports

SVG files in this directory are the **source of truth**.

To produce PNG (logo, icon, favicon, social) for platforms that require raster:

```bash
# example with rsvg-convert or inkscape if available
rsvg-convert -w 640 logo.svg -o logo.png
```

No third-party brand marks are included. Not affiliated with Railway.
