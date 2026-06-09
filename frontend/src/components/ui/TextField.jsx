import './TextField.css';

export default function TextField({
  label,
  name,
  value,
  onChange,
  type = 'text',
  placeholder,
  required,
  hint,
  autoComplete = 'off',
}) {
  return (
    <label className="field">
      <span className="field__label t-mono">
        {label}
        {required && <span className="field__req">*</span>}
      </span>
      <input
        className="field__input"
        name={name}
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        required={required}
        autoComplete={autoComplete}
        spellCheck="false"
      />
      {hint && <span className="field__hint muted">{hint}</span>}
    </label>
  );
}
