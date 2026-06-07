import { motion } from 'motion/react';
import SectionHead from '../ui/SectionHead';
import { CAPABILITIES } from '../../data/content';
import './Capabilities.css';

export default function Capabilities() {
  return (
    <section className="caps shell" id="capabilities">
      <SectionHead
        index="02 / 05"
        kicker="WHAT IT DOES"
        title="Four reflexes, one breach"
        align="right"
      />

      <div className="caps__grid">
        {CAPABILITIES.map((c, i) => (
          <motion.article
            key={c.k}
            className="cap"
            initial={{ opacity: 0, y: 26 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-50px' }}
            transition={{ duration: 0.6, delay: (i % 2) * 0.08, ease: [0.16, 1, 0.3, 1] }}
          >
            <span className="cap__index t-mono-sm muted">0{i + 1}</span>
            <span className="cap__kicker t-mono">{c.k}</span>
            <h3 className="cap__title t-display">{c.h}</h3>
            <p className="cap__text">{c.p}</p>
            <span className="cap__corner" aria-hidden="true" />
          </motion.article>
        ))}
      </div>
    </section>
  );
}
