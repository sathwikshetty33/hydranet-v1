import './Footer.css';

export default function Footer() {
  const marquee = 'TRAP THE ATTACKER — WATCH THEM WORK — CONTAIN THE BREACH — ';
  return (
    <footer className="footer">
      <div className="footer__marquee" aria-hidden="true">
        <div className="footer__track t-display">
          {Array.from({ length: 4 }).map((_, i) => (
            <span key={i}>{marquee}</span>
          ))}
        </div>
      </div>

      <div className="footer__base shell">
        <div className="footer__col">
          <span className="nav__word t-display" style={{ fontSize: '1.6rem' }}>
            Hydranet
          </span>
          <p className="muted" style={{ fontSize: 'var(--type-small)', maxWidth: '32ch', marginTop: '0.6rem' }}>
            Dynamic honeypot and incident forensics for Kubernetes. Self-hostable. Built for defenders.
          </p>
        </div>

        <div className="footer__meta t-mono-sm muted">
          <span>HYDRANET / 2026</span>
          <span>SELF-HOSTED · APACHE-2.0</span>
          <span className="footer__status">
            <span className="dot dot--live" /> SYSTEMS NOMINAL
          </span>
        </div>
      </div>
    </footer>
  );
}
