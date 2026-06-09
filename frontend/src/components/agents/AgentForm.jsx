import { useState } from 'react';
import { motion } from 'motion/react';
import Button from '../ui/Button';
import TextField from '../ui/TextField';
import './AgentForm.css';

const EMPTY = { name: '', clusterName: '' };

export default function AgentForm({ onCreate }) {
  const [form, setForm] = useState(EMPTY);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);
  const [revealed, setRevealed] = useState(null);
  const [copied, setCopied] = useState(false);

  const set = (k) => (v) => setForm((f) => ({ ...f, [k]: v }));

  async function handleSubmit(e) {
    e.preventDefault();
    if (!form.name.trim()) {
      setError('Agent name is required.');
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const created = await onCreate({
        name: form.name.trim(),
        clusterName: form.clusterName.trim(),
      });
      setRevealed(created);
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  function reset() {
    setRevealed(null);
    setForm(EMPTY);
    setCopied(false);
    setError(null);
  }

  async function copyToken() {
    if (!revealed?.auth_token) return;
    try {
      await navigator.clipboard.writeText(revealed.auth_token);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1800);
    } catch {
      // Clipboard API can fail in non-https contexts; fall through silently
      // and let the user select the text manually.
    }
  }

  if (revealed) {
    return (
      <motion.section
        className="aform"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.35, ease: [0.16, 1, 0.3, 1] }}
      >
        <header className="aform__head">
          <span className="index-marker">TOKEN ISSUED</span>
          <h2 className="aform__title t-display">Copy the token</h2>
          <p className="aform__warn">
            Shown <strong>once</strong>. Store it now — only the hash lives in the
            registry, and the raw value cannot be recovered.
          </p>
        </header>

        <div className="reveal">
          <div className="reveal__token t-mono">{revealed.auth_token}</div>
          <div className="reveal__actions">
            <button
              type="button"
              className={`reveal__copy ${copied ? 'is-copied' : ''}`}
              onClick={copyToken}
            >
              {copied ? '✓ COPIED' : '⧉ COPY'}
            </button>
            <Button onClick={reset} variant="solid">Register another</Button>
          </div>
        </div>

        <dl className="reveal__meta">
          <div>
            <dt className="t-mono muted">AGENT ID</dt>
            <dd>{String(revealed.id).padStart(3, '0')}</dd>
          </div>
          <div>
            <dt className="t-mono muted">NAME</dt>
            <dd>{revealed.name}</dd>
          </div>
          <div>
            <dt className="t-mono muted">CLUSTER</dt>
            <dd>{revealed.cluster_name || '—'}</dd>
          </div>
        </dl>

        <p className="reveal__hint muted">
          Pass via env on the agent:{' '}
          <code>HYDRANET_AGENT_ID={revealed.id}</code>{' '}
          <code>HYDRANET_AGENT_TOKEN=…</code>
        </p>
      </motion.section>
    );
  }

  return (
    <section className="aform">
      <header className="aform__head">
        <span className="index-marker">NEW AGENT</span>
        <h2 className="aform__title t-display">Register an agent</h2>
        <p className="aform__note muted">
          A fresh bearer token is generated server-side, returned <em>once</em>,
          and hashed before storage. Pass it to the agent via
          <code> HYDRANET_AGENT_TOKEN</code>.
        </p>
      </header>

      <form className="aform__form" onSubmit={handleSubmit}>
        <TextField
          label="Agent name"
          name="name"
          value={form.name}
          onChange={set('name')}
          placeholder="tetragon-agent-01"
          required
        />
        <TextField
          label="Cluster name"
          name="cluster_name"
          value={form.clusterName}
          onChange={set('clusterName')}
          placeholder="minikube"
          hint="Optional. Human-readable label for the cluster the agent runs in."
        />

        {error && (
          <p className="aform__error t-mono-sm">
            <span className="aform__error-glyph">!</span> {error}
          </p>
        )}

        <div className="aform__actions">
          <Button type="submit" variant="solid" disabled={submitting}>
            {submitting ? 'Issuing token…' : 'Register & issue token'}
          </Button>
        </div>
      </form>
    </section>
  );
}
