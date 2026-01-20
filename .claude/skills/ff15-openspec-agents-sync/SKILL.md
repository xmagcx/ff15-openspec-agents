---
name: ff15-openspec-agents-sync
description: FF15インスパイアのOpenSpecワークフロー用GitHub Copilotエージェント。Noctis（オーケストレーター + OpenSpec作成）、Iris（Issue管理）、Gladiolus（実装）、Prompto（コード品質）、Ignis（ドキュメント + アーカイブ）、Lunafreya（PR作成）を含むチーム。
---

# FF15 Copilot Agents - OpenSpec Edition

FF15チームのエージェント定義をプロジェクトに同期するスキル。

## クイックスタート

```bash
python .claude/skills/ff15-openspec-agents-sync/scripts/sync_agents.py --target .
```

## エージェント

- **Noctis** - オーケストレーター + OpenSpec作成
- **Iris** - Issue管理
- **Gladiolus** - 実装
- **Prompto** - コード品質
- **Ignis** - ドキュメント + アーカイブ
- **Lunafreya** - PR作成

## 詳細

- 使い方: [USAGE.md](USAGE.md)
- エージェント定義: `agents/*.agent.md`
- ポリシー: `docs/*-policy.md`
