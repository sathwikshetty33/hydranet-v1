import { motion } from 'motion/react';
import SectionHead from '../ui/SectionHead';
import { MODES, STACK } from '../../data/content';
import './Specs.css';

export default function Specs() {
  return (
    <section className="specs shell" id="stack">
      <SectionHead index="03 / 05" kicker="OPERATIONAL ENVELOPE" title="Adapts to your cluster" />

      <div className="specs__modes">
        {MODES.map((m, i) => (
          <motion.div
            key={m.name}
            className="mode"
            initial={{ opacity: 0, y: 24 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-50px' }}
            transition={{ duration: 0.6, delay: i * 0.08, ease: [0.16, 1, 0.3, 1] }}
          >
            <div className="mode__top">
              <span className="mode__name t-display">{m.name}</span>
              <span className="t-mono-sm muted">MODE 0{i + 1}</span>
            </div>
            <span className="mode__cni t-mono">{m.cni}</span>
            <p className="mode__detail">{m.detail}</p>
            <div className="mode__bars" aria-hidden="true">
              {Array.from({ length: 3 }).map((_, b) => (
                <span key={b} className={b <= 2 - i ? 'on' : ''} />
              ))}
            </div>
          </motion.div>
        ))}
      </div>

      <div className="specs__stack">
        <div className="specs__stack-head">
          <span className="index-marker">BUILT WITH</span>
          <p className="specs__stack-note muted">
            Go owns the cluster. Python owns the intelligence. A clean gRPC seam between them.
          </p>
        </div>
        <dl className="specs__stack-list">
          {STACK.map(([k, v]) => (
            <div className="srow" key={k}>
              <dt className="t-display">{k}</dt>
              <dd className="t-mono muted">{v}</dd>
            </div>
          ))}
        </dl>
      </div>

    </section>
  );
}
