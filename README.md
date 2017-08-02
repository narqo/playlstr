# Playlstr

At some point in future this might become a point of integration Spotify's music collection with the personal
[Feedly's Board](https://blog.feedly.com/boards/).

After an article was added to the special Board (named "Yr Next Playlist"), playlstr must parse every information
about the music album from the article (e.g. artist, title, release date) and schedule the album to be added
to Spotify collection.

```
album -> feedly board
               |-> playlstr -> queue
                                 | -> playlstr -> spotify
```
