import type { ReactNode } from 'react';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={styles.heroBanner}>
      <div className="container">
        <Heading as="h1" className="hero__title">
          {siteConfig.title}
        </Heading>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--primary button--lg"
            to="/docs/intro">
            Get Started →
          </Link>
        </div>
      </div>
    </header>
  );
}

type SectionItem = {
  title: string;
  link: string;
  description: string;
};

const sections: SectionItem[] = [
  {
    title: '📋 Getting Started',
    link: '/docs/getting-started/prerequisites',
    description: 'Prerequisites, API setup, web setup, and database configuration.',
  },
  {
    title: '🏗️ Architecture',
    link: '/docs/architecture/overview',
    description: 'System overview, authentication flow, and data models.',
  },
  {
    title: '🚀 Infrastructure',
    link: '/docs/infrastructure/ci-cd',
    description: 'CI/CD pipelines, deployment to Digital Ocean, and environment management.',
  },
  {
    title: '💻 Development',
    link: '/docs/development/api-patterns',
    description: 'API and frontend patterns, naming conventions, and migration workflow.',
  },
  {
    title: '📡 API Reference',
    link: '/docs/api/list-files',
    description: 'Endpoint contracts, request parameters, and response shapes.',
  },
  {
    title: '🤝 Contributing',
    link: '/docs/contributing/guide',
    description: 'Commit conventions, PR template, and code review process.',
  },
];

function Section({ title, link, description }: SectionItem) {
  return (
    <div className="col col--4" style={{ marginBottom: '2rem' }}>
      <Link to={link} style={{ textDecoration: 'none', color: 'inherit', display: 'block', height: '100%' }}>
        <div className="card" style={{ height: '100%', padding: '1.5rem', cursor: 'pointer', transition: 'box-shadow 0.2s' }}>
          <Heading as="h3">{title}</Heading>
          <p>{description}</p>
        </div>
      </Link>
    </div>
  );
}

export default function Home(): ReactNode {
  return (
    <Layout
      title="Home"
      description="AskAtlas documentation — architecture, infrastructure, development guides, and API reference.">
      <HomepageHeader />
      <main>
        <section style={{ padding: '2rem 0' }}>
          <div className="container">
            <div className="row">
              {sections.map((props, idx) => (
                <Section key={idx} {...props} />
              ))}
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
