<< redacted name of candidate >>
<< redacted email >>
<< redacted phone number >>
https://github.com/GabrielDCelery
<< redacted address >>

# Curriculum Vitae

## Summary

Senior Backend Engineer at a telematics insurance company. I lead backend projects end-to-end - scoping work with stakeholders, coordinating with external vendors, and guiding mid-to-junior engineers along the way. I'm hands-on with the implementation and like to work with my fellow engineers solving problems.

## Technical Expertise

| Category                 | Tools                                      |
| ------------------------ | ------------------------------------------ |
| Languages                | Typescript, Go, Python                     |
| Cloud Platforms          | AWS, Cloudflare, Azure                     |
| IaC & Config Mgmt        | CDK/Cloudformation, Terraform, Ansible     |
| CICD                     | AWS Codepipeline, CircleCI, GitHub Actions |
| Containers/Orchestration | Lambda, Docker, AWS ECS + ECR              |
| Databases/Data platforms | DynamoDB, RDS (PostgreSQL), Databricks     |
| Security & Compliance    | Trivy, Grype, Syft                         |
| Performance              | Hyperfine, Apache Benchmark                |
| Observability            | CloudWatch, X-Ray, Datadog                 |
| Messaging                | SQS, SNS, EventBridge                      |
| Source Control           | Git, GitHub                                |
| Productivity             | Neovim, Mise, AWS Vault, Zellij, Starship  |

---

## Professional Experience

### Ticker Limited (Telematics Insurance)

**Location:** United Kingdom
**Duration:** February 2020 - Present
**Roles:**

- Senior Backend Engineer (2023 - Present)
- Backend Engineer (February 2020 - March 2023)

**Key Achievements:**

- Planned and led a complete redesign and migration of our journey processing pipeline, including cost estimates and a data archival strategy
  - £150,000 yearly saving by migrating from proprietary map matching dataset to open source OSM
  - Near-real time processing of 100,000 - 300,000 journeys per day
  - Migration of 75 million (75TB) JSON files from S3 Standard storage costing ~£1400/month to ~£180/month by tarballing and zipping redundant dataset and introducing lifecycle rules to move 80% of the data to Glacier Deep Archive while retaining the ability to re-hydrate old journeys for compliance
  - Worked with data team to build data ingestion process so journeys could be analyzed in Databricks

- Extended internal customer management system coordinating 7 workflows across 63,000 policies, automating tasks for pricing, underwriting, fraud and support teams

- Built fulfilment system for 250-300 daily device shipments (sales, renewals, replacements) with automated customer notifications via email and push

- Designed near-real-time claims processing pipeline ingesting ~230 daily updates from 3 third-party providers, feeding policy management and Databricks for fraud detection

- Built billing calculator for pay-per-mile product tracking ~2,000 monthly journeys in near-real-time, integrated with external billing for automated invoicing

---

### Autologyx Ltd. (SaaS Company)

**Location:** United Kingdom
**Duration:** April 2018 - February 2020
**Role:** Full Stack Developer

Led client discovery sessions and delivered automated workflow solutions replacing manual processes.

**Key Achievements:**

- Built internal form builder tool adopted by ~20 clients for compliance workflows during frontend migration from Django to React
- Designed bespoke dashboard and end-to-end API integration for major insurance brokerage firm's reinsurance process
- Gathered requirements in client meetings, delivered automated processes with API and database integrations

---

### Arkenford Ltd. (Market Research Agency)

**Location:** United Kingdom
**Duration:** April 2016 - April 2018
**Roles:**

- Web Developer (Nov. 2016 – Apr. 2018)
- Junior Web Developer (Apr. 2016 – Nov. 2016)

Maintained internal admin tools and client-facing websites across greenfield and brownfield projects.

**Key Achievements:**

- Distributed and assisted installing ~100 telematics device in Taxis driving in London built data extraction pipeline and Leaflet heatmaps to analyze driving behaviour
- Ran surveys (500-1000 respondents) and built ETL pipeline to PostgreSQL for analysis
- Assisted migration of 200,000 customer records in PostgreSQL for compliance
- Maintained office hardware and on-prem server

---

### Freelance work

**Location:** Hungary
**Duration:** August 2015 – April 2016
**Role:** Freelance Developer

**Key Achievements:**

- Replaced Excel-based admin system (tracking 150+ customers) for one of the customers with a proper web app
- Designed and developed company website with integrated blogging functionality using external libraries

### Side projects

A collection of some of the projects I have worked on, including the languages and tools used during development: https://github.com/GabrielDCelery

- **Homelab** - Hybrid infrastructure combining k3s (cloud) and Docker (on-prem GPU), provisioned with Terraform and Ansible
- **Local AI Tools** - Document RAG and video summarization apps using Ollama and LangChain on homelab GPU
- **1 Billion Row Challenge** - TypeScript implementation focusing on performance optimization, managed to process a billion records in 10 seconds
- **Codingame Challenges** - Programming AI bots that compete against other competitors, managed to finish in the top 5% twice

### Education

Budapest University of Technology and Economics (2002-2008) / Civil Engineering
