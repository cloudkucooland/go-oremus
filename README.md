# go-oremus
##Go library to Fetch Passages from bible.oremus.org

Currently the fetch logic & parsing is not configurable. The result is an HTML block with Most formatting is removed.

## About

bible.oremus.org provides the NRSV translation of the Bible. It does not have a simple API, but the webform results are easily parsed.

Unless you need the NRSV translation, this isn't the most appropriate package for your project.

If you need a general-purpose API or other translations, there are many alternatives:

https://scripture.api.bible (no NRSV)
https://bible-api.com (no NRSV)
https://www.rkeplin.com/the-holy-bible-open-source-rest-api/ (no NRSV)
https://api.esv.org (ESV only)
https://labs.bible.org/api_web_service (no NRSV)
https://biblegateway.com (no general API) 
and countless others

## Basic Example

```
result, err := oremus.Get("Genesis 1:1-5")
if err != nil {
	//
}
fmt.Println(result)
```

## the package also includes tools to validate and normalize scripture references

CleanReference takes a string and returns a normalized string

```
result, err := oremus.CleanReference("gen 1:1-5")
if err != nil {
	//
}
fmt.Println(result)
```
Results in 
```
Genesis 1:1-5
```

ParseReferences() takes a string and returns a slice of oremus.Reference objects, these can be iterated through and printed.
```
result, err := ParseReferences("gen 1:1-5;ex 4:1,7")
if err != nil {
	//
}
for _, v := range result {
	fmt.Println(v.String())
}
```
Results in
```
Genesis 1:1-5
Exodus 4:1,7
```
