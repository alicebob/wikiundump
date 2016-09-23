wikiundump saves all Wikipedia pages from a Wikipedia XML dump to individual files.

# Usage

Download a Wikipedia dump from https://dumps.wikimedia.org/backup-index.html

  $ wikiundump -dir ./simple -keep ,Template simplewiki-*pages-articles-multistream.xml

or:

  $ bzcat simplewiki-*.xml.bz2 | wikiundump -dir ./simple -keep ,Template 


By default all pages from all namespaces are saved. `-keep` is used to store
only pages from the specified namespaces (comma separated). Note that the main
namespace is an empty string, so if you want to keep the main namespace only
use `wikiundump -keep ","` :(


# Directory structure

Files are stored with 3 levels of subdirectories, with the namespace as an extra leading subdirectory if there is any. The directory names are the first three letters of the page title, lowercased, with each non ASCII alphanum replaced with `_`. If the name of the page is shorter than 3 characters it's padded with `_`s to length 3. The title of the page will have the first letter uppercased, according to the namespace's Case (we support only "first-letter").

e.g.:

    Accordion                   -> /a/c/c/Accordion
    a cappella                  -> /a/_/c/A cappella
    Açaí Palm                   -> /a/_/a/Açaí Palm
    -1                          -> /_/1/_/-1
    A∴A∴                        -> /a/_/a/A∴A∴
    Template:Abbreviations      -> /Template/a/b/b/Template:Abbreviations
    Template:AA                 -> /Template/a/a/_/Template:AA
    Templater:AA                -> /t/e/m/Templater:AA

The list of namespaces is saved in .../namespaces.json. Later, when finding files, it
is needed to know what pages have namespace prefixes and what pages happen to have a ':' in their name. You also need to know the case rules for each namespace.


# Redirects

Pages which redirect to another page are symlinked. Use `wikiundump -symlink
false` if that's not wanted (redirected pages will be ignored).


# Filesystem

You need a case sensitive filesystem, and it must not choke on Unicode
filenames.
