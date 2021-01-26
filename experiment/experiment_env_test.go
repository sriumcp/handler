package experiment_test

import (
	"encoding/json"

	"github.com/iter8-tools/handler/experiment"
	"github.com/iter8-tools/handler/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Experiment's handler field", func() {
	Context("when containing handler actions", func() {
		var exp *experiment.Experiment
		var err error
		It("should read experiment", func() {
			exp, err = (&experiment.Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment1.yaml")).Build()
			Expect(err).Should(Succeed())
		})

		us := &unstructured.Unstructured{}
		It("should convert typed experiment into an unstructured one", func() {
			var expBytes []byte
			expBytes, err = json.Marshal(exp)
			Expect(json.Unmarshal(expBytes, &us.Object)).To(Succeed())
		})

		// It("should create the experiment", func() {
		// 	us.SetGroupVersionKind(schema.GroupVersionKind{
		// 		Group:   v2alpha1.GroupVersion.Group,
		// 		Version: v2alpha1.GroupVersion.Version,
		// 		Kind:    "Experiment",
		// 	})
		// 	log.Info("unstructured object", "us", us)
		// 	Expect(k8sClient.Create(context.Background(), us)).To(Succeed())
		// })

		// exp2 := &unstructured.Unstructured{}
		// exp2.SetGroupVersionKind(schema.GroupVersionKind{
		// 	Group:   v2alpha1.GroupVersion.Group,
		// 	Version: v2alpha1.GroupVersion.Version,
		// 	Kind:    "Experiment",
		// })
		// It("should fetch the experiment with the unknown fields", func() {
		// 	Expect(k8sClient.Get(context.Background(), types.NamespacedName{
		// 		Namespace: "default",
		// 		Name:      "exp"}, exp2)).Should(Succeed())
		// 	log.Info("fetched", "experiment", exp2)
		// 	_, found, err := unstructured.NestedFieldCopy(exp2.Object, "spec", "strategy", "handlers", "startTasks")
		// 	Expect(found).To(BeTrue())
		// 	Expect(err).To(BeNil())
		// })
	})
})
