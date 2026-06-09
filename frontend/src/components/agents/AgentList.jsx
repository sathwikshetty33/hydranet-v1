import { motion, AnimatePresence } from 'motion/react';
import Button from '../ui/Button';
import './AgentList.css';

function formatDate(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
}

function Skeleton() {
  return (
    <div className="alist__skeleton">
      {[0, 1, 2].map((i) => (
        <div key={i} className="skel-row" style={{ animationDelay: `${i * 0.12}s` }} />
      ))}
    </div>
  );
}

export default function AgentList({ agents, status, error, onRetry }) {
  if (status === 'loading') return <Skeleton />;

  if (status === 'error') {
    return (
      <div className="alist__state alist__state--err">
        <span className="t-mono flare">CONNECTION FAILED</span>
        <p className="muted">{error}</p>
        <p className="muted alist__hint">
          Is the core service up on <code>:8080</code>? Try{' '}
          <code>go run ./cmd/core</code>.
        </p>
        <Button variant="ghost" onClick={onRetry}>Retry</Button>
      </div>
    );
  }

  if (!agents.length) {
    return (
      <div className="alist__state">
        <span className="alist__empty-glyph" aria-hidden="true" />
        <span className="t-mono muted">NO AGENTS REGISTERED</span>
        <p className="muted">
          Register your first agent with the form to start receiving incidents.
        </p>
      </div>
    );
  }

  return (
    <ul className="alist">
      <AnimatePresence initial={false}>
        {agents.map((a, i) => (
          <motion.li
            key={a.id}
            className="arow"
            layout
            initial={{ opacity: 0, x: -16 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: 16, transition: { duration: 0.2 } }}
            transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
          >
            <span className="arow__idx t-mono-sm muted">
              {String(i + 1).padStart(2, '0')}
            </span>
            <div className="arow__main">
              <span className="arow__name">{a.name}</span>
              <span className="arow__cluster t-mono-sm muted">
                {a.cluster_name ? `CLUSTER · ${a.cluster_name}` : 'CLUSTER · —'}
              </span>
            </div>
            <div className="arow__meta">
              <span className="arow__date t-mono-sm muted">
                {formatDate(a.created_at)}
              </span>
              <span className="arow__status t-mono-sm">
                <span className="dot dot--live" /> ID · {String(a.id).padStart(3, '0')}
              </span>
            </div>
          </motion.li>
        ))}
      </AnimatePresence>
    </ul>
  );
}
