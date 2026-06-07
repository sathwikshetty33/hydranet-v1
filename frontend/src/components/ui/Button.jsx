import './Button.css';

export default function Button({
  children,
  to,
  href,
  variant = 'solid',
  onClick,
  type = 'button',
  disabled,
}) {
  const cls = `btn btn--${variant}`;

  if (to) {
    return (
      <a href={to} className={cls}>
        <span className="btn__label">{children}</span>
        <span className="btn__arrow" aria-hidden="true">→</span>
      </a>
    );
  }
  if (href) {
    return (
      <a href={href} className={cls} target="_blank" rel="noreferrer">
        <span className="btn__label">{children}</span>
        <span className="btn__arrow" aria-hidden="true">↗</span>
      </a>
    );
  }
  return (
    <button className={cls} onClick={onClick} type={type} disabled={disabled}>
      <span className="btn__label">{children}</span>
      <span className="btn__arrow" aria-hidden="true">→</span>
    </button>
  );
}
