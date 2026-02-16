------------------------
Bob

I've been running with sessions with Claude Code on the Web and moving between my desk and the app on my phone to provide guidance and nudging and it's been excellent

[9:32 AM]
the codexbar app is really freaking handy to know when you're running low on session tokens. Steered it to use more regular commits at each execution phase in the plan.

[9:33 AM]
But just letting agents work through a plan of actions is awesome. It's incredibly productive.

[9:34 AM]
I've actually yet to validate any of what it's produced properly, but Opus is very impressive, even though it's nomming tokens :smile:

[9:35 AM]
I can absolutely understand people not wanting to not use it and running multiple max subs across multiple things - this hobby gets expensive quickly :smile: I will be kicking gemini cli later to see what it's like - get through some Gemini tokens instead

[9:36 AM]
claude code on the web just having it installed on a single repo reduces the blast radius - it's just working on a branch or raising PRs

I feel like test coverage should just be by default. Let the robots cook and build PRs overnight to add tests (unit, integration, E2E).
Think we should be building docs including challenging scope, refactoring, leaving TASKS.md in repos so that we know what done looks like.
I'd love to have consistent doc formats across all services that can then get augmented together to identify the overall systems arch.
I'd like an agentic persona to focus on security specifically, observability etc. So that we actually go through our services with those lenses rather than just feature driven.

[4:24 PM]
walking before we can run etc, but feels like CLAUDE.md per repo building up would start compounding. Question is how much we can review PRs, how readily we can support by deploying services into ephemeral envs to run the tests automatically to validate completeness.

[4:26 PM]
there's undeniably a lot of work here, but we could and should absolutely start with some local controlled use of CC and see what some perscriptive guidance does in terms of "write to /docs only - create SECURITY.md, TESTING.md" etc. etc.)

[4:26 PM]
lots to think about

[4:27 PM]
but one thing I really wanna do is get to the point where we're not trading time with eyeballs on an IDE - it absolutely has a place, but there's real power in chunking through tokens while we're not there babysitting, knowing it can get through a boring set of tasks and report back when done with comprehensive notes. THAT to me feels like the real multiplier.

[4:28 PM]
think we could also get it to raise issues on repos as well, use that as point of activity co-ordination as well potentially


---

Andy

I don’t think reviewing notes alone gives sufficient confidence to deploy unread code into prod without significant improves to automated testing

[4:29 PM]
also - if we ever get to that stage - I don’t think the business will be able to keep up with directing what we should be writing

[4:30 PM]
nice problem to have, I guess

[4:32 PM]
I was reading an article earlier about stacked pull requests - seems like it could be a nice fit for long running agentic workflows. Spend hours working on this problem, but deliver your solution in a series of small, reviewable pull requests

---

Bob

this I think is my point - we need to rapidly get test coverage up to the point where we can be absolutely cast iron about challenge the AI to prove definition of done. Very rigid guidance in claude.md that the PR doesn't get raised until X,Y,Z is passed and metrics hit X

[4:33 PM]
sling over the article if you've still got it

[4:34 PM]
CC Web raised an initial commit to main, then started working on a branch, the presented a PR, then gave up on subsequent PRs and just commited to the branch. I ended up merging multiple into one PR

[4:34 PM]
make it work on atomic commits is key though

[4:34 PM]
much easier to reason about and keep features discrete

[4:36 PM]
Thank you for shutting me up earlier - was good to make sure that everyone had time to contribute. You know I like to talk and have every intention of starting off quie
