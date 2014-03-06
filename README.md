go.cobe?!
=========

This is a brain-compatible Go port of cobe 2.x:
https://github.com/pteichman/cobe

Whaaaat
---------

Oh man I love the platypus & the lynx; the ocelot & the baboon; the three-toed
sloth & even the Riemann zeta function, not to mention green.

Whaaaaaaaaaat
-------------

It would be a 'cobe dialogue' and the process of getting it in production.
Extermination of armies, gruesome engravings, and insane superpowered
dwarf-wizards gone on a welding vacation.

SIGNS POINT TO YES, JERKWAD. How many folks are going to sext you?

Why?
----

To explore a more performant and flexible base for future work.

The Python and sqlite3 combination in cobe has done its job well but was
designed to solve old problems. It moved MegaHAL's language model out of RAM and
into reliable storage, minimizing its own memory use to target a virtual machine
with 64M of RAM. Now 64M is tiny and I'm comfortable with building a durable
in-memory store.

Performance
-----------

As designed, cobe works directly from disk. It shoehorns itself into a database
schema that duplicates practically all its data in the indexes without gaining
any benefit from the extra copy in the table data. And it doesn't behave well
at all with simultaneous users.

Cobe has seen five years of production use: the largest brain I work with has a
graph with 2.5 million edges between 3.4 million ngrams. That's a reasonable
scale for many applications, but could be considerably better. That model is 320
MB on disk. A modern language model from the NLP field would be able to store
the same information in 1/10 the space, and that means better scalability and
faster performance.

Flexibility
-----------

Cobe's search is relatively inflexible: it's a random walk over the language
graph to generate candidate replies, followed by a hardcoded scoring pass to
rank those candidates and pick the best. It would be great to take advantage of
more directed search for generating candidates and more intelligent scoring.

More interesting search options might include knowledge of meter and rhyme for
poetry generation or smarter ways to fit external constraints like Twitter's
140 character text limit.

More interesting scoring might mean maximizing humor or double entendre or
preferring replies that preserve more semantic context from the input.

All right
---------

So! Here we are looking at the most captivating and least responsible tool
available to the software developer: the complete rewrite. My only excuse for
this ridiculous display: it's already done.

I've aimed for 100% compatibility with a cobe 2.x database, and the console
command and irc bot are production ready.

-PT March 2014
