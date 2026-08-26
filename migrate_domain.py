#!/usr/bin/env python3
"""Migrate one service domain's methods from *Client receivers to a service facade.

Usage: migrate_domain.py ServiceName file1.go [file2.go ...]

- Rewrites `func (c *Client) X(...)` to `func (s *ServiceName) X(...)`.
- Calls to migrated methods (c.X() -> s.X()) stay on the service.
- Other receiver references (c.foo -> s.client.foo) go through the Client.
- Appends transitional delegators on *Client to legacy_client.go (deleted
  once all call sites move to the facade).
"""
import re
import sys

service = sys.argv[1]
files = sys.argv[2:]

# Pass 0: collect method names defined on *Client in these files.
names = set()
for f in files:
    for m in re.finditer(r'func \(c \*Client\) (\w+)\(', open(f, encoding='utf-8').read()):
        names.add(m.group(1))
name_alt = '|'.join(sorted(names, key=len, reverse=True))

delegators = []
for f in files:
    out = []
    for line in open(f, encoding='utf-8').read().split('\n'):
        if line.lstrip().startswith('//'):
            out.append(line)
            continue
        line = re.sub(r'func \(c \*Client\) ', f'func (s *{service}) ', line)
        if names:
            line = re.sub(rf'\bc\.({name_alt})\(', r's.\1(', line)
        line = re.sub(r'\bc\.', 's.client.', line)
        out.append(line)
    src = '\n'.join(out)
    open(f, 'w', encoding='utf-8', newline='\n').write(src)

    # Collect signatures for delegators using a depth-aware scanner.
    funcs = []
    for m in re.finditer(rf'func \(s \*{service}\) (\w+)\(', src):
        name = m.group(1)
        i = m.end()  # position just past the opening '(' of params
        depth, params = 1, ''
        while depth > 0 and i < len(src):
            ch = src[i]
            if ch == '(':
                depth += 1
            elif ch == ')':
                depth -= 1
                if depth == 0:
                    break
            params += ch
            i += 1
        # return type: text between params ')' and the opening '{' at depth 0
        j = i + 1
        rets, pdepth = '', 0
        while j < len(src):
            ch = src[j]
            if ch == '(':
                pdepth += 1
            elif ch == ')':
                pdepth -= 1
            elif ch == '{' and pdepth == 0:
                break
            rets += ch
            j += 1
        funcs.append((name, params, rets.strip()))

    for name, params, rets in funcs:
        if not name[0].isupper():
            continue  # unexported helpers need no Client delegator
        params_flat = ' '.join(params.split())
        rets_flat = ' '.join(rets.split()).strip()
        # top-level comma split of params
        args, depth, cur = [], 0, ''
        for ch in params_flat:
            if ch in '([':
                depth += 1
            elif ch in ')]':
                depth -= 1
            if ch == ',' and depth == 0:
                args.append(cur)
                cur = ''
            else:
                cur += ch
        if cur.strip():
            args.append(cur)
        argnames = []
        for a in args:
            a = a.strip()
            if not a or a.startswith('func('):
                continue
            # Go params are named in implementations: the segment's first
            # identifier is the name ("ctx context.Context", grouped
            # "username, password string" splits into name-only segments).
            argnames.append(a.split()[0].lstrip('*'))
        facade = service.replace('Service', '')
        call = f'c.{facade}().{name}({", ".join(argnames)})'
        body = f'return {call}' if rets_flat else call
        sig_rets = f' {rets_flat}' if rets_flat else ''
        delegators.append(
            f'func (c *Client) {name}({params_flat}){sig_rets} {{ {body} }}\n'
        )

if delegators:
    with open('legacy_client.go', 'a', encoding='utf-8', newline='\n') as fh:
        fh.write(f'\n// --- transitional delegators: {service} ---\n')
        fh.write('\n'.join(delegators))
print(f'migrated {len(names)} methods onto {service}, {len(delegators)} delegators appended')
