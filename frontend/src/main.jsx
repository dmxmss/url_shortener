import React, { useMemo, useState } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

const config = window.__URL_SHORTENER_CONFIG__ || {};
const apiBaseUrl = (config.API_BASE_URL || "").replace(/\/$/, "");

function normalizeUrl(value) {
  const trimmed = value.trim();
  if (!trimmed) {
    return "";
  }

  if (/^https?:\/\//i.test(trimmed)) {
    return trimmed;
  }

  return `https://${trimmed}`;
}

function App() {
  const [url, setUrl] = useState("");
  const [result, setResult] = useState(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);

  const canSubmit = useMemo(() => normalizeUrl(url).length > 0 && !loading, [url, loading]);

  async function handleSubmit(event) {
    event.preventDefault();
    setError("");
    setResult(null);
    setCopied(false);

    const normalized = normalizeUrl(url);
    if (!normalized) {
      setError("Введите URL для сокращения.");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch(`${apiBaseUrl}/api/shorten`, {
        method: "POST",
        headers: {
          "content-type": "application/json"
        },
        body: JSON.stringify({ url: normalized })
      });

      const payload = await response.json().catch(() => ({}));
      if (!response.ok) {
        throw new Error(payload.error || "Не удалось создать короткую ссылку.");
      }

      setResult(payload);
    } catch (err) {
      setError(err.message || "Не удалось подключиться к API.");
    } finally {
      setLoading(false);
    }
  }

  async function copyResult() {
    if (!result?.short_url) {
      return;
    }

    await navigator.clipboard.writeText(result.short_url);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 1600);
  }

  return (
    <main className="shell">
      <section className="panel" aria-labelledby="title">
        <div className="heading">
          <p className="eyebrow">URL Shortener</p>
          <h1 id="title">Сократить ссылку</h1>
        </div>

        <form className="form" onSubmit={handleSubmit}>
          <label htmlFor="url">Длинная ссылка</label>
          <div className="inputRow">
            <input
              id="url"
              type="text"
              inputMode="url"
              placeholder="https://example.com/very/long/url"
              value={url}
              onChange={(event) => setUrl(event.target.value)}
              autoComplete="url"
            />
            <button type="submit" disabled={!canSubmit}>
              {loading ? "..." : "Сократить"}
            </button>
          </div>
        </form>

        {error && <p className="message error">{error}</p>}

        {result && (
          <div className="result" role="status">
            <div>
              <span>Короткая ссылка</span>
              <a href={result.short_url} target="_blank" rel="noreferrer">
                {result.short_url}
              </a>
            </div>
            <button className="secondary" type="button" onClick={copyResult}>
              {copied ? "Скопировано" : "Копировать"}
            </button>
          </div>
        )}
      </section>
    </main>
  );
}

createRoot(document.getElementById("root")).render(<App />);
