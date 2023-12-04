# Instructions

```
cd cmd/aimakeme
go build
EXPORT OPENAI_APIKEY="..."
./aimakeme -p "Something that makes me smile" -f happy
cd ~/aimakeme/happy
ls
```

# How to Use
Flags:
- p - The prompt to use
- f - The sub-directory in ~/aimakeme to store these images. Use this when you wanna keep using the same or similar prompt, and keep the photos together
- s - The style, which is `natural` or `vivid`
- n - The number of images to generate. These requests will be made in parallel