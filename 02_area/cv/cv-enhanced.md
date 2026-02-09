<< redacted name of candidate >>
<< redacted email >>
<< redacted phone number >>
https://github.com/GabrielDCelery
<< redacted address >>

# Curriculum Vitae

## Summary

Senior Backend Engineer at a telematics insurance company. I lead backend projects end-to-end - scoping work with stakeholders, coordinating with external vendors, and guiding mid-to-junior engineers along the way. I'm hands-on with the implementation and like to work with my fellow engineers solving problems.

## Technical Expertise

| Category                 | Tools                                            |
| ------------------------ | ------------------------------------------------ |
| Languages                | Typescript, Go, SQL (PostgreSQL, SQLite), Python |
| Cloud Platforms          | AWS, Cloudflare, Azure                           |
| IaC & Config Mgmt        | CDK/Cloudformation, Terraform, Ansible           |
| CICD                     | AWS Codepipeline, CircleCI, GitHub Actions       |
| Containers/Orchestration | Lambda, Docker, AWS ECS + ECR                    |
| Security & Compliance    | Trivy, Grype, Syft                               |
| Performance              | Hyperfine, Apache Benchmark                      |
| Observability            | CloudWatch, X-Ray, Datadog                       |
| Messaging                | SQS, SNS, EventBridge                            |
| AI Integration & Tooling | Claude API, GitHub Copilot                       |
| Source Control           | Git, GitHub                                      |
| Productivity             | Neovim, Mise, AWS Vault, Zellij, Starship        |

---

## Professional Experience

### Ticker Limited (Telematics Insurance)

**Location:** United Kingdom
**Duration:** February 2020 - Present
**Roles:**

- Senior Backend Engineer (2023 - Present)
- Backend Engineer (February 2020 - March 2023)

**Key Achievements:**

- Extended and enhanced our internal customer management system that coordinated 7 different workflows across 63,000 insurance policies used by our pricing, underwriting, fraud and support teams to automate repetitive work

- Designed and built fulfilment system that handled the issuing, tracking and labeling of daily 250-300 devices at the time of policy sale, renewal and faulty device replacements and integrated it with our email and push notification system so the customers were kept up-to-date on the status of their device

- Designed and built a near-real time claims processing system that ingested daily ~230 updates to claims from three different third-party providers and forwarded them to the policy management system and Databricks for pattern recognition and fraud prevention

- Built the billing calculator and tracker for our pay-per-mile product - calculated costs for ~2,000 monthly journeys in near-real time and synced with external billing system for monthly bill generation

- Planned and led a complete redesign of our journey processing pipeline migration, including cost estimates and a data archival strategy
  - £150,000 yearly saving by migrating from proprietary map matching dataset to open source OSM
  - Near-real time processing of 100,000 - 300,000 journeys per day
  - Migration of 75 million (75TB) JSON files from S3 Standard costing ~£1400/month to ~£180/month by tarballing and zipping redundant dataset and introducing lifecycle rule to move 80% of the data to Glacier Deep Archive while retaining the ability to re-hydrate old journeys for compliance
  - worked with data team to build data ingestion process so journeys could be analyzed in Databricks

---

### Autologyx Ltd. (SaaS Company)

**Location:** United Kingdom
**Duration:** April 2018 - February 2020
**Role:** Full Stack Developer

Worked with clients to turn their manual processes into automated workflows. Ran discovery sessions to understand what they needed, then built and shipped the solution.

**Key Achievements:**

- Designed and built a temporary internal form builder tool that was used across ~20 clients to drive compliance processes while our main engineering team finished migrating our product's frontent from Django to React
- Designed and built a bespoke dashboard and end-to-end API integration for one of the largest independent insurance brokerages to handle one of their re-insurance process
- Attended to business meetings and discoveries to understand client needs and built automated processes, API and database integrations using the company's software

---

### Arkenford Ltd. (Market Research Agency)

**Location:** United Kingdom
**Duration:** April 2016 - April 2018
**Roles:**

- Developer (Nov. 2016 – Apr. 2018)
- Junior Web Developer (Apr. 2016 – Nov. 2016)

Maintained internal admin tools and client websites. Worked on both new builds and existing systems.

**Key Achievements:**

- Distributed and assisted installing ~100 telematics device in Taxis driving in London and after collecting the devices wrote the script to extract the driving data and plot them using Leaflet to analyze heatmap data of driving behaviour to derive insights
- Created and ran regulary surveys across 500-1000 respondents and then extracted responses into internal PostgreSQL database for analysis
- Assisted migration of PostgreSQL dataset of 200000 customer records for compliance reasons
- Maintenance of office computers and server hosted in the office

---

### Freelance work

**Location:** Hungary
**Duration:** August 2015 – April 2016
**Role:** Freelance Developer

**Key Achievements:**

- Replaced Excel-based admin system (tracking 150+ customers) for one of the customers with a proper web app
- Talked to clients and users to figure out what they actually needed
- Built the full-stack app - frontend and backend
- Tested the UI with real users and fixed what didn't work
- Built their website and a simple CMS for posting news updates
- Worked with the client to make sure it did what they needed
- Added a blog feature using off-the-shelf libraries

### Side projects

A collection of some of the projects I have worked on, including the languages and tools used during development: https://github.com/GabrielDCelery

- **Homelab** - Kubernetes cluster with Terraform/Ansible provisioning, hybrid cloud-on-prem networking using DigitalOcean
- **1 Billion Row Challenge** - TypeScript implementation focusing on performance optimization, managed to process a billion records in 10 seconds
- **Codingame Challenges** - Programming AI bots that compete against other competitors, managed to finish in the top 5% twice

### Education

Budapest University of Technology and Economics (2002-2008) / Civil Engineering
