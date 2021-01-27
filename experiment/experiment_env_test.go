package experiment_test

import (
	"context"
	"encoding/json"

	"github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/handler/experiment"
	"github.com/iter8-tools/handler/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Experiment's handler field", func() {
	Context("when containing handler actions", func() {
		var exp *experiment.Experiment
		var err error
		It("should deal with the handler actions properly", func() {
			By("reading the experiment from file")
			exp, err = (&experiment.Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment1.yaml")).Build()
			Expect(err).ToNot(HaveOccurred())

			By("converting the type experiment into an unstructured one")
			us := &unstructured.Unstructured{}
			var expBytes []byte
			expBytes, err = json.Marshal(exp)
			Expect(json.Unmarshal(expBytes, &us.Object)).To(Succeed())

			By("k8s creating experiment in cluster")
			us.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   v2alpha1.GroupVersion.Group,
				Version: v2alpha1.GroupVersion.Version,
				Kind:    "Experiment",
			})
			Expect(k8sClient.Create(context.Background(), us)).To(Succeed())

			By("fetching experiment from cluster")
			b := &experiment.Builder{}
			exp2, err := b.FromCluster("sklearn-iris-experiment-1", "default", k8sClient).Build()
			Expect(err).ToNot(HaveOccurred())

			Expect(exp2.Spec).To(Equal(exp.Spec))
		})
	})
})
