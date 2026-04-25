import { Extension } from "@tiptap/core";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import { Decoration, DecorationSet } from "@tiptap/pm/view";

/**
 * TipTap extension that visually highlights the selection an AI edit
 * is operating on. Uses a ProseMirror Decoration (inline class) so
 * the highlight moves with the text on scroll, resize, or doc edits
 * -- no fixed overlay div + window-event listener pile of duct tape.
 *
 * The bubble menu drives this via `setMeta(aiSelectionPluginKey, ...)`:
 *   - On "Ask AI" click: { from, to, status: "idle" }
 *   - When stream starts:  { from, to, status: "streaming" }
 *   - On close / cancel:   null
 */

export type AiSelectionStatus = "idle" | "streaming";

export interface AiSelectionState {
  from: number;
  to: number;
  status: AiSelectionStatus;
}

export const aiSelectionPluginKey = new PluginKey<AiSelectionState | null>(
  "aiSelectionRange",
);

const HIGHLIGHT_CLASS = "ai-edit-range";
const STREAMING_CLASS = "ai-edit-range--streaming";

export const AiSelectionRange = Extension.create({
  name: "aiSelectionRange",
  addProseMirrorPlugins() {
    return [
      new Plugin<AiSelectionState | null>({
        key: aiSelectionPluginKey,
        state: {
          init: () => null,
          apply(tr, value) {
            const meta = tr.getMeta(aiSelectionPluginKey);
            if (meta !== undefined) {
              return meta as AiSelectionState | null;
            }
            if (!value) return value;
            // Map the highlighted range through any doc changes so it
            // stays anchored to the same text the user selected.
            if (tr.docChanged) {
              const from = tr.mapping.map(value.from);
              const to = tr.mapping.map(value.to);
              if (to <= from) return null;
              return { ...value, from, to };
            }
            return value;
          },
        },
        props: {
          decorations(state) {
            const value = aiSelectionPluginKey.getState(state);
            if (!value) return null;
            const { from, to, status } = value;
            if (from >= to) return null;
            const className =
              status === "streaming"
                ? `${HIGHLIGHT_CLASS} ${STREAMING_CLASS}`
                : HIGHLIGHT_CLASS;
            return DecorationSet.create(state.doc, [
              Decoration.inline(from, to, { class: className }),
            ]);
          },
        },
      }),
    ];
  },
});
