import { motion } from 'motion/react';
import './SectionHead.css';

export default function SectionHead({ index, kicker, title, align = 'left' }) {
  return (
    <motion.div
      className={`sec-head sec-head--${align}`}
      initial={{ opacity: 0, y: 24 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: '-80px' }}
      transition={{ duration: 0.7, ease: [0.16, 1, 0.3, 1] }}
    >
      <div className="sec-head__meta">
        <span className="index-marker">{index}</span>
        <span className="t-mono muted">{kicker}</span>
      </div>
      <h2 className="sec-head__title t-display">{title}</h2>
    </motion.div>
  );
}
