import { visit } from "unist-util-visit";
import type { Plugin } from "unified";
import type { Root, Nodes, Parent } from "mdast";
import type {
  ContainerDirective,
  LeafDirective,
  TextDirective,
} from "mdast-util-directive";

import { ENTITY_DIRECTIVE_NAMES, type EntityType } from "./extract-refs";

const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

type AnyDirective = ContainerDirective | LeafDirective | TextDirective;

function isEntityName(name: string): name is EntityType {
  return (ENTITY_DIRECTIVE_NAMES as readonly string[]).includes(name);
}

function setEntityHast(node: AnyDirective, inline: boolean): boolean {
  const id = node.attributes?.id;
  if (typeof id !== "string" || !UUID_RE.test(id)) {
    return false;
  }
  const data = node.data ?? (node.data = {});
  data.hName = `ask-${node.name}-ref`;
  data.hProperties = {
    id: id.toLowerCase(),
    "data-inline": inline ? "1" : "0",
  };
  return true;
}

function setCalloutHast(node: ContainerDirective) {
  const data = node.data ?? (node.data = {});
  data.hName = "ask-callout";
  data.hProperties = {
    "data-callout-type": node.attributes?.type ?? "note",
  };
}

function rawFor(node: AnyDirective): string {
  const attrs = node.attributes ?? {};
  const attrStr = Object.entries(attrs)
    .map(([k, v]) => `${k}="${v ?? ""}"`)
    .join(" ");
  const prefix = node.type === "containerDirective" ? ":::" : "::";
  return attrStr ? `${prefix}${node.name}{${attrStr}}` : `${prefix}${node.name}`;
}

function replaceWithText(parent: Parent, index: number, raw: string) {
  parent.children.splice(index, 1, { type: "text", value: raw });
}

export const remarkAskAtlasDirectives: Plugin<[], Root> = () => (tree) => {
  visit(tree, (node: Nodes, index, parent) => {
    if (
      node.type !== "textDirective" &&
      node.type !== "leafDirective" &&
      node.type !== "containerDirective"
    ) {
      return;
    }
    const directive = node as AnyDirective;

    if (isEntityName(directive.name)) {
      const ok = setEntityHast(directive, directive.type === "textDirective");
      if (!ok && parent && typeof index === "number") {
        replaceWithText(parent, index, rawFor(directive));
      }
      return;
    }

    if (
      directive.name === "callout" &&
      directive.type === "containerDirective"
    ) {
      setCalloutHast(directive);
      return;
    }

    if (parent && typeof index === "number") {
      replaceWithText(parent, index, rawFor(directive));
    }
  });
};
