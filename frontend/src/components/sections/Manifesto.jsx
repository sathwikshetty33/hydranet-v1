import { motion } from 'motion/react';
import './Manifesto.css';

const lines = [
  { t: 'Kill the pod', kind: 'strike' },
  { t: 'and you kill the evidence.', kind: 'plain' },
  { t: 'The intrusion vector,', kind: 'plain' },
  { t: 'the lateral map,', kind: 'plain' },
  { t: 'the intent —', kind: 'plain' },
  { t: 'gone in one kubectl delete.', kind: 'mute' },
];

export default function Manifesto() {
  return (
    <section className="manifesto shell" id="manifesto">
      <p className="manifesto__tag t-mono">
        <span className="dot dot--warn" /> THE PROBLEM WITH STANDARD IR
      </p>

      <h2 className="manifesto__statement t-display">
        {lines.map((l, i) => (
          <motion.span
            key={i}
            className={`manifesto__line manifesto__line--${l.kind}`}
            initial={{ opacity: 0.06, y: 18 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-15%' }}
            transition={{ duration: 0.6, delay: i * 0.08, ease: [0.16, 1, 0.3, 1] }}
          >
            {l.t}{' '}
          </motion.span>
        ))}
      </h2>

      <motion.p
        className="manifesto__resolve"
        initial={{ opacity: 0, y: 20 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.7, delay: 0.2 }}
      >
        So HydraNet refuses to delete. It quarantines the pod where it stands, lets the
        attacker keep digging, and turns the breach into the richest intelligence source
        you have — <span className="blood">a honeypot made from your own infrastructure.</span>
      </motion.p>
    </section>
  );
}
