//
// Copyright 2016 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package policy

import (
	"encoding/json"

	"github.com/cilium/cilium/common"
	"github.com/cilium/cilium/pkg/labels"

	. "gopkg.in/check.v1"
)

type CommonSuite struct{}

var _ = Suite(&CommonSuite{})

func (s *CommonSuite) TestReservedID(c *C) {
	i1 := labels.GetID("host")
	c.Assert(i1, Equals, labels.ID_HOST)
	c.Assert(i1.String(), Equals, "host")

	i2 := labels.GetID("world")
	c.Assert(i2, Equals, labels.ID_WORLD)
	c.Assert(i2.String(), Equals, "world")

	c.Assert(labels.GetID("unknown"), Equals, labels.ID_UNKNOWN)
	unknown := labels.ReservedID(700)
	c.Assert(unknown.String(), Equals, "")
}

func (s *CommonSuite) TestUnmarshalAllowRule(c *C) {
	var rule AllowRule

	longLabel := `{"source": "kubernetes", "key": "!io.kubernetes.pod.name", "value": "foo"}`
	invLabel := `{"source": "kubernetes", "value": "foo"}`
	shortLabel := `"web"`
	invertedLabel := `"!web"`

	err := json.Unmarshal([]byte(longLabel), &rule)
	c.Assert(err, Equals, nil)
	c.Assert(rule.Action, Equals, DENY)
	c.Assert(rule.Label.Source, Equals, "kubernetes")
	c.Assert(rule.Label.AbsoluteKey(), Equals, "io.kubernetes.pod.name")
	c.Assert(rule.Label.Value, Equals, "foo")

	err = json.Unmarshal([]byte(invLabel), &rule)
	c.Assert(err, Not(Equals), nil)

	err = json.Unmarshal([]byte(shortLabel), &rule)
	c.Assert(err, Equals, nil)
	c.Assert(rule.Action, Equals, ACCEPT)
	c.Assert(rule.Label.Source, Equals, common.CiliumLabelSource)
	c.Assert(rule.Label.AbsoluteKey(), Equals, "web")
	c.Assert(rule.Label.Value, Equals, "")

	err = json.Unmarshal([]byte(invertedLabel), &rule)
	c.Assert(err, Equals, nil)
	c.Assert(rule.Action, Equals, DENY)
	c.Assert(rule.Label.Source, Equals, common.CiliumLabelSource)
	c.Assert(rule.Label.AbsoluteKey(), Equals, "web")
	c.Assert(rule.Label.Value, Equals, "")

	err = json.Unmarshal([]byte(""), &rule)
	c.Assert(err, Not(Equals), nil)
}

func (s *CommonSuite) TestNodeCovers(c *C) {
	root := Node{
		Name: common.GlobalLabelPrefix,
		Children: map[string]*Node{
			"foo": {},
			"bar": {},
		},
	}

	err := root.ResolveTree()
	c.Assert(err, Equals, nil)

	lblFoo := labels.NewLabel("io.cilium.foo", "", common.CiliumLabelSource)
	ctx := SearchContext{To: []labels.Label{*lblFoo}}
	c.Assert(root.Covers(&ctx), Equals, true)
	c.Assert(root.Children["foo"].Covers(&ctx), Equals, true)
	c.Assert(root.Children["bar"].Covers(&ctx), Equals, false)

	lblFoo = labels.NewLabel("io.cilium.foo2", "", common.CiliumLabelSource)
	ctx = SearchContext{To: []labels.Label{*lblFoo}}
	c.Assert(root.Covers(&ctx), Equals, true)
	c.Assert(root.Children["foo"].Covers(&ctx), Equals, false)
	c.Assert(root.Children["bar"].Covers(&ctx), Equals, false)

	lblRoot := labels.NewLabel("io.cilium", "", common.CiliumLabelSource)
	ctx = SearchContext{To: []labels.Label{*lblRoot}}
	c.Assert(root.Covers(&ctx), Equals, true)
	c.Assert(root.Children["foo"].Covers(&ctx), Equals, false)
	c.Assert(root.Children["bar"].Covers(&ctx), Equals, false)
}

func (s *CommonSuite) TestLabelCompare(c *C) {
	a1 := labels.NewLabel("io.cilium", "", "")
	a2 := labels.NewLabel("io.cilium", "", "")
	b1 := labels.NewLabel("io.cilium.bar", "", common.CiliumLabelSource)
	c1 := labels.NewLabel("io.cilium.bar", "", "kubernetes")
	d1 := labels.NewLabel("", "", "")

	c.Assert(a1.Equals(a2), Equals, true)
	c.Assert(a2.Equals(a1), Equals, true)
	c.Assert(a1.Equals(b1), Equals, false)
	c.Assert(a1.Equals(c1), Equals, false)
	c.Assert(a1.Equals(d1), Equals, false)
	c.Assert(b1.Equals(c1), Equals, false)
}

func (s *CommonSuite) TestAllowRule(c *C) {
	lblFoo := labels.NewLabel("io.cilium.foo", "", common.CiliumLabelSource)
	lblBar := labels.NewLabel("io.cilium.bar", "", common.CiliumLabelSource)
	lblBaz := labels.NewLabel("io.cilium.baz", "", common.CiliumLabelSource)
	lblAll := labels.NewLabel(labels.ID_NAME_ALL, "", common.ReservedLabelSource)
	allow := AllowRule{Action: ACCEPT, Label: *lblFoo}
	deny := AllowRule{Action: DENY, Label: *lblFoo}
	allowAll := AllowRule{Action: ACCEPT, Label: *lblAll}

	ctx := SearchContext{
		From: []labels.Label{*lblFoo},
		To:   []labels.Label{*lblBar},
	}
	ctx2 := SearchContext{
		From: []labels.Label{*lblBaz},
		To:   []labels.Label{*lblBar},
	}

	c.Assert(allow.Allows(&ctx), Equals, ACCEPT)
	c.Assert(deny.Allows(&ctx), Equals, DENY)
	c.Assert(allowAll.Allows(&ctx), Equals, ACCEPT)
	c.Assert(allow.Allows(&ctx2), Equals, UNDECIDED)
	c.Assert(deny.Allows(&ctx2), Equals, UNDECIDED)
	c.Assert(allowAll.Allows(&ctx2), Equals, ACCEPT)
}

func (s *CommonSuite) TestTargetCoveredBy(c *C) {
	lblFoo := labels.NewLabel("io.cilium.foo", "", common.CiliumLabelSource)
	lblBar := labels.NewLabel("io.cilium.bar", "", common.CiliumLabelSource)
	lblBaz := labels.NewLabel("io.cilium.baz", "", common.CiliumLabelSource)
	lblJoe := labels.NewLabel("io.cilium.user", "joe", "kubernetes")
	lblAll := labels.NewLabel(labels.ID_NAME_ALL, "", common.ReservedLabelSource)

	list1 := []labels.Label{*lblFoo}
	list2 := []labels.Label{*lblBar, *lblBaz}
	list3 := []labels.Label{*lblFoo, *lblJoe}
	list4 := []labels.Label{*lblAll}

	// any -> io.cilium.bar
	ctx := SearchContext{To: []labels.Label{*lblBar}}
	c.Assert(ctx.TargetCoveredBy(list1), Equals, false)
	c.Assert(ctx.TargetCoveredBy(list2), Equals, true)
	c.Assert(ctx.TargetCoveredBy(list3), Equals, false)
	c.Assert(ctx.TargetCoveredBy(list4), Equals, true)

	// any -> kubernetes:io.cilium.baz
	ctx = SearchContext{To: []labels.Label{*lblBaz}}
	c.Assert(ctx.TargetCoveredBy(list1), Equals, false)
	c.Assert(ctx.TargetCoveredBy(list2), Equals, true)
	c.Assert(ctx.TargetCoveredBy(list3), Equals, false)
	c.Assert(ctx.TargetCoveredBy(list4), Equals, true)

	// any -> [kubernetes:io.cilium.user=joe, io.cilium.foo]
	ctx = SearchContext{To: []labels.Label{*lblJoe, *lblFoo}}
	c.Assert(ctx.TargetCoveredBy(list1), Equals, true)
	c.Assert(ctx.TargetCoveredBy(list2), Equals, false)
	c.Assert(ctx.TargetCoveredBy(list3), Equals, true)
	c.Assert(ctx.TargetCoveredBy(list4), Equals, true)
}

func (s *CommonSuite) TestAllowConsumer(c *C) {
	lblTeamA := labels.NewLabel("io.cilium.teamA", "", common.CiliumLabelSource)
	lblTeamB := labels.NewLabel("io.cilium.teamB", "", common.CiliumLabelSource)
	lblFoo := labels.NewLabel("io.cilium.foo", "", common.CiliumLabelSource)
	lblBar := labels.NewLabel("io.cilium.bar", "", common.CiliumLabelSource)
	lblBaz := labels.NewLabel("io.cilium.baz", "", common.CiliumLabelSource)

	// [Foo,TeamA] -> Bar
	aFooToBar := SearchContext{
		From: []labels.Label{*lblTeamA, *lblFoo},
		To:   []labels.Label{*lblBar},
	}

	// [Baz, TeamA] -> Bar
	aBazToBar := SearchContext{
		From: []labels.Label{*lblTeamA, *lblBaz},
		To:   []labels.Label{*lblBar},
	}

	// [Foo,TeamB] -> Bar
	bFooToBar := SearchContext{
		From: []labels.Label{*lblTeamB, *lblFoo},
		To:   []labels.Label{*lblBar},
	}

	// [Baz, TeamB] -> Bar
	bBazToBar := SearchContext{
		From: []labels.Label{*lblTeamB, *lblBaz},
		To:   []labels.Label{*lblBar},
	}

	allowFoo := AllowRule{Action: ACCEPT, Label: *lblFoo}
	dontAllowFoo := AllowRule{Action: DENY, Label: *lblFoo}
	allowTeamA := AllowRule{Action: ACCEPT, Label: *lblTeamA}
	dontAllowBaz := AllowRule{Action: DENY, Label: *lblBaz}
	alwaysAllowFoo := AllowRule{Action: ALWAYS_ACCEPT, Label: *lblFoo}

	// Allow: foo, !foo
	consumers := PolicyRuleConsumers{
		Coverage: []labels.Label{*lblBar},
		Allow:    []AllowRule{allowFoo, dontAllowFoo},
	}

	// NOTE: We are testing on single consumer rule leve, there is
	// no default deny policy enforced. No match equals UNDECIDED

	c.Assert(consumers.Allows(&aFooToBar), Equals, DENY)
	c.Assert(consumers.Allows(&bFooToBar), Equals, DENY)
	c.Assert(consumers.Allows(&aBazToBar), Equals, UNDECIDED)
	c.Assert(consumers.Allows(&bBazToBar), Equals, UNDECIDED)

	// Always-Allow: foo, !foo
	consumers = PolicyRuleConsumers{
		Coverage: []labels.Label{*lblBar},
		Allow:    []AllowRule{alwaysAllowFoo, dontAllowFoo},
	}

	c.Assert(consumers.Allows(&aFooToBar), Equals, ALWAYS_ACCEPT)
	c.Assert(consumers.Allows(&bFooToBar), Equals, ALWAYS_ACCEPT)
	c.Assert(consumers.Allows(&aBazToBar), Equals, UNDECIDED)
	c.Assert(consumers.Allows(&bBazToBar), Equals, UNDECIDED)

	// Allow: TeamA, !baz
	consumers = PolicyRuleConsumers{
		Coverage: []labels.Label{*lblBar},
		Allow:    []AllowRule{allowTeamA, dontAllowBaz},
	}

	c.Assert(consumers.Allows(&aFooToBar), Equals, ACCEPT)
	c.Assert(consumers.Allows(&aBazToBar), Equals, DENY)
	c.Assert(consumers.Allows(&bFooToBar), Equals, UNDECIDED)
	c.Assert(consumers.Allows(&bBazToBar), Equals, DENY)

	// Allow: TeamA, !baz
	consumers = PolicyRuleConsumers{
		Coverage: []labels.Label{*lblFoo},
		Allow:    []AllowRule{allowTeamA, dontAllowBaz},
	}

	c.Assert(consumers.Allows(&aFooToBar), Equals, UNDECIDED)
	c.Assert(consumers.Allows(&aBazToBar), Equals, UNDECIDED)
	c.Assert(consumers.Allows(&bFooToBar), Equals, UNDECIDED)
	c.Assert(consumers.Allows(&bBazToBar), Equals, UNDECIDED)
}

func (s *CommonSuite) TestBuildPath(c *C) {
	rootNode := Node{Name: common.GlobalLabelPrefix}
	p, err := rootNode.BuildPath()
	c.Assert(p, Equals, common.GlobalLabelPrefix)
	c.Assert(err, Equals, nil)

	// missing parent assignment
	fooNode := Node{Name: "foo"}
	p, err = fooNode.BuildPath()
	c.Assert(p, Equals, "")
	c.Assert(err, Not(Equals), nil)

	rootNode.Children = map[string]*Node{"foo": &fooNode}
	fooNode.Parent = &rootNode
	p, err = fooNode.BuildPath()
	c.Assert(p, Equals, common.GlobalLabelPrefix+".foo")
	c.Assert(err, Equals, nil)

	err = rootNode.ResolveTree()
	c.Assert(err, Equals, nil)
	c.Assert(rootNode.path, Equals, common.GlobalLabelPrefix)
	c.Assert(fooNode.path, Equals, common.GlobalLabelPrefix+".foo")

}

func (s *CommonSuite) TestValidateCoverage(c *C) {
	rootNode := Node{Name: common.GlobalLabelPrefix}
	node := Node{
		Name:   "foo",
		Parent: &rootNode,
	}

	lblBar := labels.NewLabel("io.cilium.bar", "", common.CiliumLabelSource)
	consumer := PolicyRuleConsumers{Coverage: []labels.Label{*lblBar}}
	c.Assert(consumer.Resolve(&node), Not(Equals), nil)

	consumer2 := PolicyRuleRequires{Coverage: []labels.Label{*lblBar}}
	c.Assert(consumer2.Resolve(&node), Not(Equals), nil)

	lblFoo := labels.NewLabel("io.cilium.foo", "", common.CiliumLabelSource)
	consumer = PolicyRuleConsumers{Coverage: []labels.Label{*lblFoo}}
	c.Assert(consumer.Resolve(&node), Equals, nil)

	lblFoo = labels.NewLabel("foo", "", common.CiliumLabelSource)
	consumer = PolicyRuleConsumers{Coverage: []labels.Label{*lblFoo}}
	c.Assert(consumer.Resolve(&node), Equals, nil)
}

func (s *CommonSuite) TestRequires(c *C) {
	lblFoo := labels.NewLabel("io.cilium.foo", "", common.CiliumLabelSource)
	lblBar := labels.NewLabel("io.cilium.bar", "", common.CiliumLabelSource)
	lblBaz := labels.NewLabel("io.cilium.baz", "", common.CiliumLabelSource)

	// Foo -> Bar
	aFooToBar := SearchContext{
		From: []labels.Label{*lblFoo},
		To:   []labels.Label{*lblBar},
	}

	// Baz -> Bar
	aBazToBar := SearchContext{
		From: []labels.Label{*lblBaz},
		To:   []labels.Label{*lblBar},
	}

	// Bar -> Baz
	aBarToBaz := SearchContext{
		From: []labels.Label{*lblBar},
		To:   []labels.Label{*lblBaz},
	}

	// coverage: bar
	// Require: foo
	requires := PolicyRuleRequires{
		Coverage: []labels.Label{*lblBar},
		Requires: []labels.Label{*lblFoo},
	}

	c.Assert(requires.Allows(&aFooToBar), Equals, UNDECIDED)
	c.Assert(requires.Allows(&aBazToBar), Equals, DENY)
	c.Assert(requires.Allows(&aBarToBaz), Equals, UNDECIDED)
}

func (s *CommonSuite) TestNodeAllows(c *C) {
	lblProd := labels.NewLabel("io.cilium.Prod", "", common.CiliumLabelSource)
	lblQA := labels.NewLabel("io.cilium.QA", "", common.CiliumLabelSource)
	lblFoo := labels.NewLabel("io.cilium.foo", "", common.CiliumLabelSource)
	lblBar := labels.NewLabel("io.cilium.bar", "", common.CiliumLabelSource)
	lblBaz := labels.NewLabel("io.cilium.baz", "", common.CiliumLabelSource)
	lblJoe := labels.NewLabel("io.cilium.user", "joe", "kubernetes")
	lblPete := labels.NewLabel("io.cilium.user", "pete", "kubernetes")

	// [Foo,QA] -> [Bar,QA]
	qaFooToQaBar := SearchContext{
		From: []labels.Label{*lblQA, *lblFoo},
		To:   []labels.Label{*lblBar, *lblQA},
	}

	// [Foo, Prod] -> [Bar,Prod]
	prodFooToProdBar := SearchContext{
		From: []labels.Label{*lblProd, *lblFoo},
		To:   []labels.Label{*lblBar},
	}

	// [Foo,QA] -> [Bar,prod]
	qaFooToProdBar := SearchContext{
		From: []labels.Label{*lblQA, *lblFoo},
		To:   []labels.Label{*lblBar, *lblProd},
	}

	// [Foo,QA, Joe] -> [Bar,prod]
	qaJoeFooToProdBar := SearchContext{
		From: []labels.Label{*lblQA, *lblFoo, *lblJoe},
		To:   []labels.Label{*lblBar, *lblProd},
	}

	// [Foo,QA, Pete] -> [Bar,Prod]
	qaPeteFooToProdBar := SearchContext{
		From: []labels.Label{*lblQA, *lblFoo, *lblPete},
		To:   []labels.Label{*lblBar, *lblProd},
	}

	// [Baz, QA] -> Bar
	qaBazToQaBar := SearchContext{
		From: []labels.Label{*lblQA, *lblBaz},
		To:   []labels.Label{*lblQA, *lblBar},
	}

	rootNode := Node{
		Name: common.GlobalLabelPrefix,
		Rules: []PolicyRule{
			&PolicyRuleConsumers{
				Coverage: []labels.Label{*lblBar},
				Allow: []AllowRule{
					{ // always-allow:  user=joe
						Action: ALWAYS_ACCEPT,
						Label:  *lblJoe,
					},
					{ // allow:  user=pete
						Action: ACCEPT,
						Label:  *lblPete,
					},
				},
			},
			&PolicyRuleRequires{ // coverage qa, requires qa
				Coverage: []labels.Label{*lblQA},
				Requires: []labels.Label{*lblQA},
			},
			&PolicyRuleRequires{ // coverage prod, requires: prod
				Coverage: []labels.Label{*lblProd},
				Requires: []labels.Label{*lblProd},
			},
			&PolicyRuleConsumers{
				Coverage: []labels.Label{*lblBar},
				Allow: []AllowRule{
					{ // allow: foo
						Action: ACCEPT,
						Label:  *lblFoo,
					},
				},
			},
		},
	}

	c.Assert(rootNode.ResolveTree(), Equals, nil)

	c.Assert(rootNode.Allows(&qaFooToQaBar), Equals, ACCEPT)
	c.Assert(rootNode.Allows(&prodFooToProdBar), Equals, ACCEPT)
	c.Assert(rootNode.Allows(&qaFooToProdBar), Equals, DENY)
	c.Assert(rootNode.Allows(&qaJoeFooToProdBar), Equals, ALWAYS_ACCEPT)
	c.Assert(rootNode.Allows(&qaPeteFooToProdBar), Equals, DENY)
	c.Assert(rootNode.Allows(&qaBazToQaBar), Equals, UNDECIDED)
}

func (s *CommonSuite) TestResolveTree(c *C) {
	rootNode := Node{
		Name: common.GlobalLabelPrefix,
		Children: map[string]*Node{
			"foo": {Rules: []PolicyRule{&PolicyRuleConsumers{}}},
		},
	}

	c.Assert(rootNode.ResolveTree(), Equals, nil)
	c.Assert(rootNode.Children["foo"].Name, Equals, "foo")
}

func (s *CommonSuite) TestpolicyAllows(c *C) {
	lblProd := labels.NewLabel("io.cilium.Prod", "", common.CiliumLabelSource)
	lblQA := labels.NewLabel("io.cilium.QA", "", common.CiliumLabelSource)
	lblFoo := labels.NewLabel("io.cilium.foo", "", common.CiliumLabelSource)
	lblBar := labels.NewLabel("io.cilium.bar", "", common.CiliumLabelSource)
	lblBaz := labels.NewLabel("io.cilium.baz", "", common.CiliumLabelSource)
	lblJoe := labels.NewLabel("io.cilium.user", "joe", "kubernetes")
	lblPete := labels.NewLabel("io.cilium.user", "pete", "kubernetes")

	// [Foo,QA] -> [Bar,QA]
	qaFooToQaBar := SearchContext{
		From: []labels.Label{*lblQA, *lblFoo},
		To:   []labels.Label{*lblQA, *lblBar},
	}

	// [Foo, Prod] -> [Bar,Prod]
	prodFooToProdBar := SearchContext{
		From: []labels.Label{*lblProd, *lblFoo},
		To:   []labels.Label{*lblBar},
	}

	// [Foo,QA] -> [Bar,Prod]
	qaFooToProdBar := SearchContext{
		From: []labels.Label{*lblQA, *lblFoo},
		To:   []labels.Label{*lblBar, *lblProd},
	}

	// [Foo,QA, Joe] -> [Bar,prod]
	qaJoeFooToProdBar := SearchContext{
		From: []labels.Label{*lblQA, *lblFoo, *lblJoe},
		To:   []labels.Label{*lblBar, *lblProd},
	}

	// [Foo,QA, Pete] -> [Bar,Prod]
	qaPeteFooToProdBar := SearchContext{
		From: []labels.Label{*lblQA, *lblFoo, *lblPete},
		To:   []labels.Label{*lblBar, *lblProd},
	}

	// [Baz, QA] -> Bar
	qaBazToQaBar := SearchContext{
		From: []labels.Label{*lblQA, *lblBaz},
		To:   []labels.Label{*lblQA, *lblBar},
	}

	rootNode := Node{
		Name: common.GlobalLabelPrefix,
		Rules: []PolicyRule{
			&PolicyRuleConsumers{
				Coverage: []labels.Label{*lblBar},
				Allow: []AllowRule{
					// always-allow: user=joe
					{Action: ALWAYS_ACCEPT, Label: *lblJoe},
					// allow:  user=pete
					{Action: ACCEPT, Label: *lblPete},
				},
			},
			&PolicyRuleRequires{ // coverage qa, requires qa
				Coverage: []labels.Label{*lblQA},
				Requires: []labels.Label{*lblQA},
			},
			&PolicyRuleRequires{ // coverage prod, requires: prod
				Coverage: []labels.Label{*lblProd},
				Requires: []labels.Label{*lblProd},
			},
		},
		Children: map[string]*Node{
			"foo": {},
			"bar": {
				Rules: []PolicyRule{
					&PolicyRuleConsumers{
						Allow: []AllowRule{
							{ // allow: foo
								Action: ACCEPT,
								Label:  *lblFoo,
							},
							{Action: DENY, Label: *lblJoe},
							{Action: DENY, Label: *lblPete},
						},
					},
				},
			},
		},
	}

	c.Assert(rootNode.ResolveTree(), Equals, nil)

	root := Tree{&rootNode}
	c.Assert(root.Allows(&qaFooToQaBar), Equals, ACCEPT)
	c.Assert(root.Allows(&prodFooToProdBar), Equals, ACCEPT)
	c.Assert(root.Allows(&qaFooToProdBar), Equals, DENY)
	c.Assert(root.Allows(&qaJoeFooToProdBar), Equals, ACCEPT)
	c.Assert(root.Allows(&qaPeteFooToProdBar), Equals, DENY)
	c.Assert(root.Allows(&qaBazToQaBar), Equals, DENY)

	_, err := json.MarshalIndent(rootNode, "", "    ")
	c.Assert(err, Equals, nil)
}

func (s *CommonSuite) TestNodeMerge(c *C) {
	// Name mismatch
	aNode := Node{Name: "a"}
	bNode := Node{Name: "b"}
	err := aNode.Merge(&bNode)
	c.Assert(err, Not(Equals), nil)

	// Empty nodes
	aOrig := Node{Name: "a"}
	aNode = Node{Name: "a"}
	bNode = Node{Name: "a"}
	err = aNode.Merge(&bNode)
	c.Assert(err, Equals, nil)
	c.Assert(aNode, DeepEquals, aOrig)

	lblProd := labels.NewLabel("io.cilium.Prod", "", common.CiliumLabelSource)
	lblQA := labels.NewLabel("io.cilium.QA", "", common.CiliumLabelSource)
	lblFoo := labels.NewLabel("io.cilium.foo", "", common.CiliumLabelSource)
	lblJoe := labels.NewLabel("io.cilium.user", "joe", "kubernetes")
	lblPete := labels.NewLabel("io.cilium.user", "pete", "kubernetes")

	aNode = Node{
		Name: common.GlobalLabelPrefix,
		Rules: []PolicyRule{
			&PolicyRuleRequires{ // coverage qa, requires qa
				Coverage: []labels.Label{*lblQA},
				Requires: []labels.Label{*lblQA},
			},
		},
		Children: map[string]*Node{
			"bar": {
				Name: "bar",
				path: common.GlobalLabelPrefix + ".bar",
				Rules: []PolicyRule{
					&PolicyRuleConsumers{
						Allow: []AllowRule{
							{Action: DENY, Label: *lblJoe},
							{Action: DENY, Label: *lblPete},
						},
					},
				},
			},
		},
	}

	bNode = Node{
		Name: common.GlobalLabelPrefix,
		Rules: []PolicyRule{
			&PolicyRuleRequires{ // coverage prod, requires: prod
				Coverage: []labels.Label{*lblProd},
				Requires: []labels.Label{*lblProd},
			},
		},
		Children: map[string]*Node{
			"foo": {
				Name: "foo",
				path: common.GlobalLabelPrefix + ".foo",
			},
			"bar": {
				Name: "bar",
				path: common.GlobalLabelPrefix + ".bar",
				Rules: []PolicyRule{
					&PolicyRuleConsumers{
						Allow: []AllowRule{
							{ // allow: foo
								Action: ACCEPT,
								Label:  *lblFoo,
							},
						},
					},
				},
			},
		},
	}

	aNode.Path()
	bNode.Path()

	err = aNode.Merge(&bNode)
	c.Assert(err, Equals, nil)
}

func (s *CommonSuite) TestSearchContextReplyJSON(c *C) {
	scr := SearchContextReply{
		Logging:  []byte(`foo`),
		Decision: ConsumableDecision(0x1),
	}
	scrWanted := SearchContextReply{
		Logging:  []byte(`foo`),
		Decision: ConsumableDecision(0x1),
	}
	b, err := json.Marshal(scr)
	c.Assert(err, IsNil)
	c.Assert(b, DeepEquals, []byte(`{"Logging":"Zm9v","Decision":"accept"}`))

	var scrGot SearchContextReply
	err = json.Unmarshal(b, &scrGot)
	c.Assert(err, IsNil)
	c.Assert(scrGot, DeepEquals, scrWanted)
}
