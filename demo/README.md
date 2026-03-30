# Demo Assets

Prepare the demo repository and render the terminal GIF:

```bash
./demo/render-demo.sh
```

This writes the rendered asset to `assets/envguard-demo.gif`.

The visible tape is generated from `demo/envguard.tape.tmpl` so the demo can inline the real computed fingerprint without hardcoding a stale value.
