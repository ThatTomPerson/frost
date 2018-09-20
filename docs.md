you have two `Manifest` classes. `InstalledManifest` and `LockedManifest`

One represents what is installed and what the end should look like 

a manifest is a tree of `Module`s

making a `Diff` from the two manifests will produce a `ActionList`
which contains `Delete` or `Install` `Actions`

<!-- `Update` is only for git dependancies -->

<!-- `Install` pushes a job to a `DownloadQueue` which tracks what has been downloaded and saves it to a cache -->

`Install` Has three steps
 - `Download` which downloads to cache or clones a repo to a tmp dir
 - `Install` which copies from the cache to the dist dir
 - `PostInstall` which is used to simlink stuff to bin dir or compile c

Each handler has a `Cache` implimentation
most will just use the `VersionedModuleCache` which uses the module's name and version as the cache key