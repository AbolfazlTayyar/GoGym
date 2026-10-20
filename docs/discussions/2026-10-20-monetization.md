# Monetization — 2026-10-20

> **Status:** a record of one conversation, not a decision. The answers are what was true on
> this date. Read it alongside [product-direction.md](../product-direction.md), which it builds
> on.

**Question asked:** do you have any recommendation for monetizing this app? Before you answer,
you can grill me with questions.

Before asking anything, Claude read [product-direction.md](../product-direction.md) and
[stakeholder-brief.md](../stakeholder-brief.md), so the questions skip what was already decided
(not selling to gyms, athletes get a link rather than an app, metering by athlete count).

## The questions and answers

### Round 1

| Question | Answer |
|---|---|
| Where does the app stand with the one coach right now? | **Not using it yet.** v1 is still being built and the notebook hasn't moved in. |
| Who do you picture paying? | **All four:** other coaches (SaaS), athletes through coaches (a cut of plan sales), athletes directly, and third parties (supplement shops, gyms, brands). |
| Is the market Iran, with Iranian payment rails (Zarinpal/IDPay, no recurring card billing)? | **Not decided.** |
| What are you aiming for, and how much time can you put into it? | **Side income:** some money each month from a few hours a week, mostly hands-off. |

### Round 2

| Question | Answer |
|---|---|
| How does your coach get public (online) athletes, and roughly how many a month? | **Rarely.** The gym is the business. |
| How would the first 10 paying coaches hear about it? | **The coach's network, Instagram/Telegram, and shared plan links.** |
| Will you personally onboard and support coaches? | **The coach could do it**, as a partner or ambassador. |
| What do coaches you know pay for today to run their work? | **Nothing:** notebook, free tools, WhatsApp. |

### Round 3

| Question | Answer |
|---|---|
| How many coaches could your coach realistically introduce you to? | **Under 5.** |
| What's your relationship with the coach if this makes money? | "We will make an agreement, I trust her." Not formalized yet. |
| Have you looked at what Iranian coaches already use for this? | Nothing good exists, and coaches find it hard to pay for organizing tools. Many prefer handwritten plans. |

## The recommendation

### Don't build anything for monetization yet

The coach isn't using the app yet. Coaches around her pay for nothing today, and many prefer
handwritten plans. Earlier tools failed, and she can introduce fewer than 5 coaches. A billing
system, or any design for one, would be built for customers who don't exist yet. The first goal
is a coach who would be upset to lose the app. Money comes after that.

### 1. Make the one coach depend on it (now through v1)

- **The bar:** she uses it daily in the gym for 4 weeks without being pushed. If she drifts back
  to the notebook, no pricing model will fix it.
- **Measure two numbers:** how long a plan takes to write (the target is under 15 minutes, down
  from about an hour) and how many plans she builds each week. These numbers are also the sales
  pitch later.
- **Do her notebook import with her.** Typing in 20–30 plans on the first day is where
  handwritten-plan coaches give up. Making that easy will win coaches more than any feature.

### 2. Build the share link next

`/p/<token>` is already first in the "sellable" list in
[product-direction.md](../product-direction.md#sequencing), and these answers make it more
important:

- It's how other coaches hear about the app, and right now it's the only one that scales. Each
  plan a coach sends carries "made with GoGym" to athletes, and some of those athletes are
  coaches.
- It's the one benefit a coach who pays nothing will notice: she looks more professional than a
  coach with a notebook or a Canva PDF.

### 3. Give it free to her 3–4 coaches, then ask them to pay

- Give it away first and watch whether they're still using it after a month without nudging.
  That tells you more than any survey.
- Then charge one simple price: free up to about 3 athletes, paid above that. Sell **prepaid
  periods** (3, 6 or 12 months), not monthly. That fits Iranian payments, and a coach who pays
  for a year doesn't quietly cancel.
- **Collect payment by hand at first:** a card-to-card transfer, then extend the coach's access
  with `coachctl`. With fewer than 10 coaches, a payment gateway, invoices and subscription
  tables are exactly the machinery [CLAUDE.md](../../CLAUDE.md) warns against. A `paid_until`
  column can wait until it's needed, because adding it later is one migration with no backfill.
  Getting a gateway like Zarinpal usually needs a registered business and an eNamad trust seal,
  so that's another reason to wait until manual renewals become a chore.
- **Agree on the coach's share now**, while it costs nothing. A simple version is a percentage
  of what each coach she brings in pays, for a year. That gives her a reason to sell the app to
  other coaches.

### Dropped for now

- **A cut of athlete payments.** She rarely sells online, and handling money for other people on
  Iranian payment rails is a legal and payment-processing project on its own.
- **Athletes paying directly.** It needs athlete accounts, and athletes already get a link
  instead.
- **Third parties (supplement shops, brands).** They only pay for reach, and a reach of tens of
  athletes is worth nothing to them.
- **Iran vs. diaspora.** Left undecided. Nothing to be built in the next six months depends on
  it.

### What to expect

With fewer than 5 coaches to reach directly and a market that pays for nothing today, this will
be small side income at best for the first year. The pitch that can work on a coach who pays
nothing is "it gives you your evenings back and makes you look professional", proven with her
real numbers. The larger opportunity, if it ever comes, is helping coaches sell online plans
through the share link. She doesn't do that much today, so it stays a guess until a coach asks
for it.

## Open after this discussion

- The agreement with the coach: her share, and for how long.
- Iran only, or Persian-speaking coaches abroad too.
- What prepaid price the first 3–4 coaches will accept, tested once they've used it free for a
  month.
