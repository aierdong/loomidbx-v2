# Agentic UI DSL / Semantic IR 指导守则

# 一、核心目标（最重要）

DSL 的目标不是“描述页面”。

而是：

> 用最少 token 表达生成 UI 所必须知道的“结构意图”。

记住：

## 不是：

```txt
页面长什么样
```

## 而是：

```txt
页面承担什么职责
```

---

# 二、DSL 的核心哲学（必须牢记）

---

# 原则 1：表达 Intent，而不是 Appearance

## 好：

```yaml
role: auth-form
purpose: login
```

## 不好：

```yaml
title: 登录
color: blue
padding: 24
```

---

# 原则 2：表达 Semantic Structure，而不是 DOM

## 好：

```yaml
sections:
  - hero
  - navigation-hub
  - activity-feed
```

## 不好：

```yaml
div:
  div:
    div:
```

---

# 原则 3：表达 UI Responsibility，而不是视觉细节

## 好：

```yaml
role: primary-navigation
```

## 不好：

```yaml
two cards with icon on top
```

---

# 原则 4：Agent 需要的是“意图”，不是“文案”

文案：

* 可以后生成
* 可以替换
* 可以 i18n

而：

* layout
* interaction
* hierarchy
* role

才是真正决定代码结构的。

---

# 原则 5：压缩“重复信息”

## 好：

```yaml
card-grid:
  item: project-card
  count: 12
```

## 不好：

```yaml
card1
card2
card3
...
```

---

# 三、DSL 的真正职责

DSL 不应该成为：

```txt
HTML 的另一种写法
```

而应该成为：

## “UI Intent Graph”

---

# 四、推荐的 DSL 层级结构（标准化）

推荐统一：

```yaml
page:
layout:
sections:
components:
interactions:
state:
data:
responsive:
accessibility:
navigation:
```

这是一个非常稳定的 schema。

---

# 五、最推荐的字段（重点）

---

# 1. role（最重要）

这是 DSL 的核心。

## 推荐：

```yaml
role: auth-form
role: navigation-hub
role: activity-feed
role: settings-panel
```

role 决定：

* agent prior
* layout 推断
* component 推断
* interaction 推断

---

# 2. purpose（第二重要）

## 推荐：

```yaml
purpose: user-login
purpose: feature-discovery
purpose: project-navigation
```

purpose 是业务意图。

---

# 3. layout intent

不要写 CSS。

写：

```yaml
layout: split-screen
layout: centered-column
layout: grid(3)
```

---

# 4. interaction intent

不要写：

```yaml
onclick
```

写：

```yaml
intent: navigate-to-project
intent: toggle-theme
intent: submit-auth
```

---

# 5. responsive intent

必须表达。

因为 agent 很容易乱猜。

## 推荐：

```yaml
responsive:
  mobile: stacked
  desktop: split-screen
```

---

# 6. state semantics

高级但非常重要。

## 推荐：

```yaml
state:
  theme:
    values: [light, dark]
    persistent: true
```

---

# 7. data semantics

非常重要。

## 推荐：

```yaml
data:
  source: recent-projects
  type: collection
  mutable: true
```

---

# 六、最重要的设计原则：有限 Vocabulary

这是未来你 Skill 最关键的一部分。

---

# 不要无限发明 role

## 不好：

```yaml
role: magical-project-explorer
```

## 好：

```yaml
role: navigation-hub
```

---

# 建立 Canonical Ontology

推荐维护：

```txt
auth-form
navigation-hub
marketing-panel
activity-feed
entity-list
data-table
editor-shell
settings-panel
dashboard-grid
empty-state
command-bar
modal-dialog
wizard-flow
```

这是极其重要的。

因为：

LLM 对固定 ontology 的推断能力远强于自由文本。

---

# 七、真正应该删除的东西

---

# 不要保留：

## 1. CSS

```yaml
padding
margin
color
font-size
```

全部删除。

---

# 2. DOM 结构

```yaml
div:
  div:
```

完全没意义。

---

# 3. SVG/path

无意义 token 污染。

---

# 4. 精确视觉

```yaml
width: 734px
```

不要。

---

# 5. 重复内容

不要列举：

```yaml
card1
card2
card3
```

而要：

```yaml
item: project-card
count: 3
```

---

# 八、什么时候应该保留文案

只有：

## 文案本身影响结构时

例如：

```yaml
cta-priority: destructive
```

或者：

```yaml
empty-state-message
```

否则：

大部分 text 都可以省略。

---

# 九、Agent 最需要什么

这是整个守则最重要的一句：

> Agent 最需要的是“结构确定性”。

不是：

* 视觉 fidelity
* 文案 fidelity

而是：

## “生成时不要猜”

---

# 十、优秀 DSL 的标准

好的 DSL：

## 1. token 极少

## 2. 信息密度极高

## 3. 可预测

## 4. 可组合

## 5. 可推断

## 6. 可复用

---

# 十一、你的 DSL 不应该做什么

DSL：

## 不是设计稿

## 不是 HTML

## 不是 Figma JSON

## 不是 React AST

---

# 十二、DSL 的真正定位

DSL 应该：

```txt
高于 HTML
低于 Product Spec
```

处于：

## “UI Intent Layer”

---

# 十三、推荐的 Agent Pipeline（非常推荐）

这是你未来可以直接做成 Skill 的。

```txt
HTML / Screenshot
        ↓
DOM Cleaner
        ↓
Semantic Extractor
        ↓
Canonical Role Mapping
        ↓
UI DSL / Semantic IR
        ↓
Agent Codegen
```

---

# 十四、非常关键：Canonicalization

你未来一定会遇到：

```yaml
role: project-list
role: projects
role: recent-projects
role: project-feed
```

这会毁掉稳定性。

所以必须：

## 统一 vocabulary

例如：

```txt
project-list
```

永远只用一个。

---

# 十五、真正高级的方向（未来）

你未来甚至可以：

---

# 1. 建立 UI Ontology

类似：

```txt
Navigation
Data Display
Input
Feedback
Workflow
```

---

# 2. 建立 Layout Grammar

例如：

```txt
split-screen
stack
grid
master-detail
sidebar-shell
```

---

# 3. 建立 Interaction Taxonomy

例如：

```txt
navigate
submit
toggle
expand
filter
sort
```

---

# 十六、最后的黄金法则（请记住）

---

# 黄金法则 1

## 不要描述“是什么样”

要描述：

## “为什么存在”

---

# 黄金法则 2

## 不要表达视觉

要表达：

## responsibility

---

# 黄金法则 3

## 不要给 agent 看代码

要给：

## intent graph

---

# 黄金法则 4

## 不要保留 implementation

要保留：

## semantics

---

# 黄金法则 5（最重要）

一个 DSL 字段存在的唯一理由是：

> “如果删掉它，agent 会更容易生成错误代码。”

否则：

删掉它。
