// @tags: [does_not_support_stepdowns]

t = db.getCollection("basic1");
t.drop();

o = {
    a: 1
};
t.insertOne(o);

assert.eq(1, t.findOne().a, "first");
assert(o._id, "now had id");
assert(o._id.str, "id not a real id");

o.a = 2;
t.insertOne(o);

assert.eq(2, t.findOne().a, "second");

// not a very good test of currentOp, but tests that it at least
// is sort of there:
assert(db.currentOp().inprog != null);
