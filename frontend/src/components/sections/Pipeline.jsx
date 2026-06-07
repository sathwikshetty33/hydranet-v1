import { motion } from 'motion/react';
import SectionHead from '../ui/SectionHead';
import { PIPELINE } from '../../data/content';
import './Pipeline.css';

export default function Pipeline() {
  return (
    <section className="pipeline shell" id="how">
      <SectionHead index="01 / 05" kicker="INCIDENT LIFECYCLE" title="From alarm to autopsy" />

      <ol className="pipeline__list">
        {PIPELINE.map((step, i) => (
          <motion.li
            key={step.id}
            className="pstep"
            initial={{ opacity: 0, y: 30 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-60px' }}
            transition={{ duration: 0.6, delay: i * 0.05, ease: [0.16, 1, 0.3, 1] }}
          >
            <span className="pstep__num t-display">{step.id}</span>
            <div className="pstep__body">
              <div className="pstep__head">
                <h3 className="pstep__title t-display">{step.title}</h3>
                <span className="pstep__tech t-mono">{step.technique}</span>
              </div>
              <p className="pstep__text">{step.body}</p>
            </div>
            <span className="pstep__rule" aria-hidden="true" />
          </motion.li>
        ))}
      </ol>
    </section>
  );
}
