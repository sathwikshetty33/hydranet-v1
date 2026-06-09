import { useCallback, useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { listIncidents } from '../api/incidents';
import Button from '../components/ui/Button';
import './Incidents.css';

const REFRESH_INTERVAL_MS = 10_000;

function formatStamp(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '—';
  return d.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  });
}

function severityClass(sev) {
  switch ((sev || '').toLowerCase()) {
    case 'critical': return 'is-crit';
    case 'high':     return 'is-high';
    case 'medium':   return 'is-warn';
    case 'low':      return 'is-low';
    default:         return '';
  }
}

function Skeleton() {
  return (
    <div className="ilist__skeleton">
      {[0, 1, 2, 3].map((i) => (
        <div key={i} className="skel-row" style={{ animationDelay: `${i * 0.12}s` }} />
      ))}
    </div>
  );
}

export default function Incidents() {
  const [incidents, setIncidents] = useState([]);
  const [status, setStatus] = useState('loading');
  const [error, setError] = useState(null);
  const [refreshedAt, setRefreshedAt] = useState(null);

  const reload = useCallback(async ({ background } = {}) => {
    if (!background) setStatus('loading');
    setError(null);
    try {
      const rows = await listIncidents();
      setIncidents(rows);
      setStatus('ready');
      setRefreshedAt(new Date());
    } catch (e) {
      setError(e.message);
      setStatus('error');
    }
  }, []);

  // Initial fetch
  useEffect(() => {
    let active = true;
    (async () => {
      try {
        const rows = await listIncidents();
        if (!active) return;
        setIncidents(rows);
        setStatus('ready');
        setRefreshedAt(new Date());
      } catch (e) {
        if (!active) return;
        setError(e.message);
        setStatus('error');
      }
    })();
    return () => { active = false; };
  }, []);

  // Auto-refresh in the background — silent on success so the UI doesn't
  // flash; visible if it newly fails.
  useEffect(() => {
    const id = window.setInterval(() => {
      reload({ background: true });
    }, REFRESH_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [reload]);

  return (
    <main className="incidents">
      <header className="incidents__head shell">
        <div className="incidents__title-block">
          <span className="index-marker">CONSOLE // INCIDENT LOG</span>
          <h1 className="incidents__title t-display">Incidents</h1>
        </div>
        <div className="incidents__readout t-mono">
          <span className="incidents__readout-row">
            <span className="muted">RECORDED</span>
            <span className="incidents__count">
              {String(incidents.length).padStart(3, '0')}
            </span>
          </span>
          <span className="incidents__readout-row">
            <span className="muted">LAST SYNC</span>
            <span className="incidents__sync">
              {refreshedAt ? formatStamp(refreshedAt.toISOString()) : '—'}
            </span>
          </span>
        </div>
      </header>

      <div className="divider" />

      <div className="incidents__body shell">
        <div className="incidents__toolbar">
          <span className="t-mono muted">AUTO-REFRESH · {REFRESH_INTERVAL_MS / 1000}s</span>
          <Button onClick={() => reload()} variant="ghost">Refresh now</Button>
        </div>

        {status === 'loading' && incidents.length === 0 && <Skeleton />}

        {status === 'error' && (
          <div className="ilist__state ilist__state--err">
            <span className="t-mono flare">CONNECTION FAILED</span>
            <p className="muted">{error}</p>
            <p className="muted ilist__hint">
              Is the core service up on <code>:8080</code>?
            </p>
            <Button variant="ghost" onClick={() => reload()}>Retry</Button>
          </div>
        )}

        {status === 'ready' && incidents.length === 0 && (
          <div className="ilist__state">
            <span className="ilist__empty-glyph" aria-hidden="true" />
            <span className="t-mono muted">NO INCIDENTS YET</span>
            <p className="muted">
              Register an agent and trigger a suspicious exec; entries land here
              as soon as the agent reports.
            </p>
          </div>
        )}

        {status === 'ready' && incidents.length > 0 && (
          <ul className="ilist">
            <AnimatePresence initial={false}>
              {incidents.map((inc) => (
                <motion.li
                  key={inc.id}
                  className="irow"
                  layout
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.3, ease: [0.16, 1, 0.3, 1] }}
                >
                  <div className="irow__top">
                    <span className="irow__id t-mono">
                      INC-{String(inc.id).padStart(4, '0')}
                    </span>
                    <span className={`irow__sev ${severityClass(inc.severity)} t-mono-sm`}>
                      {(inc.severity || 'unknown').toUpperCase()}
                    </span>
                    <span className="irow__stamp t-mono-sm muted">
                      {formatStamp(inc.created_at)}
                    </span>
                  </div>
                  <p className="irow__msg">{inc.message || '—'}</p>
                  <div className="irow__meta t-mono-sm">
                    <span className="irow__chip">
                      <span className="muted">AGENT</span>{' '}
                      {inc.agent_name || `#${inc.agent_id}`}
                    </span>
                    <span className="irow__chip">
                      <span className="muted">CLUSTER</span>{' '}
                      {inc.cluster_name || '—'}
                    </span>
                    <span className="irow__chip">
                      <span className="muted">POD</span> {inc.pod_name || '—'}
                    </span>
                    <span className="irow__chip">
                      <span className="muted">SVC</span> {inc.service_name || '—'}
                    </span>
                  </div>
                </motion.li>
              ))}
            </AnimatePresence>
          </ul>
        )}
      </div>
    </main>
  );
}
