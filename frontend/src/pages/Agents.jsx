import { useCallback, useEffect, useState } from 'react';
import { listAgents, createAgent } from '../api/agents';
import AgentForm from '../components/agents/AgentForm';
import AgentList from '../components/agents/AgentList';
import './Agents.css';

export default function Agents() {
  const [agents, setAgents] = useState([]);
  const [status, setStatus] = useState('loading');
  const [error, setError] = useState(null);

  const reload = useCallback(async () => {
    setStatus('loading');
    setError(null);
    try {
      setAgents(await listAgents());
      setStatus('ready');
    } catch (e) {
      setError(e.message);
      setStatus('error');
    }
  }, []);

  useEffect(() => {
    let active = true;
    (async () => {
      try {
        const data = await listAgents();
        if (!active) return;
        setAgents(data);
        setStatus('ready');
      } catch (e) {
        if (!active) return;
        setError(e.message);
        setStatus('error');
      }
    })();
    return () => {
      active = false;
    };
  }, []);

  const handleCreate = useCallback(async (input) => {
    const created = await createAgent(input);
    // Prepend; the list endpoint is ID-DESC so this stays consistent until reload.
    setAgents((prev) => [created, ...prev]);
    return created;
  }, []);

  return (
    <main className="agents">
      <header className="agents__head shell">
        <div className="agents__title-block">
          <span className="index-marker">CONSOLE // AGENT REGISTRY</span>
          <h1 className="agents__title t-display">Agents</h1>
        </div>
        <div className="agents__readout t-mono">
          <span className="agents__readout-row">
            <span className="muted">REGISTERED</span>
            <span className="agents__count">
              {status === 'ready' ? String(agents.length).padStart(2, '0') : '--'}
            </span>
          </span>
          <span className="agents__readout-row">
            <span className="muted">CORE</span>
            <span className={`agents__link agents__link--${status}`}>
              <span className={`dot ${status === 'error' ? 'dot--warn' : 'dot--live'}`} />
              {status === 'error' ? 'OFFLINE' : status === 'loading' ? 'LINKING' : 'LINKED'}
            </span>
          </span>
        </div>
      </header>

      <div className="divider" />

      <div className="agents__body shell">
        <div className="agents__col agents__col--form">
          <AgentForm onCreate={handleCreate} />
        </div>
        <div className="agents__col agents__col--list">
          <div className="agents__list-head">
            <span className="t-mono muted">ACTIVE REGISTRY</span>
            <span className="t-mono-sm muted">NAME · CLUSTER · ADDED</span>
          </div>
          <AgentList
            agents={agents}
            status={status}
            error={error}
            onRetry={reload}
          />
        </div>
      </div>
    </main>
  );
}
