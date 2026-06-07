import { lazy, Suspense } from 'react';
import Button from '../ui/Button';
import './Hero.css';

const HeroCanvas = lazy(() => import('../three/HeroCanvas'));

export default function Hero() {
  return (
    <section className="hero scanlines">
      <div className="hero__canvas">
        <Suspense fallback={<div className="hero__canvas-fallback" />}>
          <HeroCanvas />
        </Suspense>
      </div>

      <div className="hero__grid shell">
        <div className="hero__top rise rise-d-1">
          <span className="index-marker">SEC // K8S HONEYPOT</span>
          <span className="t-mono muted hero__coord">34.05°N · 118.24°W · NODE-01</span>
        </div>

        <h1 className="hero__title t-display">
          <span className="rise rise-d-2">Hydra</span>
          <span className="hero__title-accent flicker rise rise-d-3">net</span>
        </h1>

        <div className="hero__sub rise rise-d-4">
          <p className="hero__lede">
            When a pod is breached, the reflex is to kill it — and burn the evidence with it.
            HydraNet does the opposite. It <em>seals the attacker in</em>, keeps the session
            alive, and watches every move.
          </p>
          <div className="hero__actions">
            <Button href="https://github.com" variant="solid">Read the docs</Button>
            <Button to="#how" variant="ghost">How it works</Button>
          </div>
        </div>

        <div className="hero__ribbon t-mono-sm rise rise-d-5">
          <span><span className="dot" /> ISOLATE IN-PLACE</span>
          <span className="hero__ribbon-sep">/</span>
          <span>ENUMERATE SECRETS</span>
          <span className="hero__ribbon-sep">/</span>
          <span>LLM FORENSICS</span>
          <span className="hero__ribbon-sep">/</span>
          <span>ESCAPE INTERLOCK</span>
        </div>
      </div>

      <div className="hero__scroll t-mono-sm muted" aria-hidden="true">
        SCROLL ↓
      </div>
    </section>
  );
}
