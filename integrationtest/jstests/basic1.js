// @tags: [does_not_support_stepdowns]

t = db.getCollection("basic1");
t.drop();

o = {
    a: 1
};
t.insertOne(o);
assert(1==t.findOne().a, "first")

t.updateOne({ a: 1 }, { $set: { a: 2 } })
assert(2==t.findOne().a, "second")

// not a very good test of currentOp, but tests that it at least
// is sort of there:
assert(db.currentOp().inprog != null);
